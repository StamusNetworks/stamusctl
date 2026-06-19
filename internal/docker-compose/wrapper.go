package compose

import (
	// Core
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync/atomic"

	// Common
	"stamus-ctl/internal/backup"
	stamusFlags "stamus-ctl/internal/handlers"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/shutdown"
	"stamus-ctl/internal/stamus"
	"stamus-ctl/internal/utils"

	// External
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/cmd/compatibility"
	commands "github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// operationCounter is used to generate unique operation IDs.
var operationCounter atomic.Int64

// Constants
var ComposeFlags = models.ComposeFlags{
	"up": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{"detach", "build", "remove-orphans"},
	),
	"down": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{"volumes", "remove-orphans"},
	),
	"restart": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{},
	),
	"exec": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{"detach", "privileged", "user", "workdir", "env", "no-TTY", "dry-run", "index"},
	),
	"ps": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{"services", "quiet", "format"},
	),
	"logs": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{"timestamps", "tail", "since", "until", "follow", "details"},
	),
	"pull": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{"ignore-buildable", "ignore-pull-failures", "include-deps", "quiet"},
	),
	"images": models.CreateComposeFlags(
		[]string{"file", "project-name"},
		[]string{"format", "quiet"},
	),
}

// Variables
// var ComposeCmds map[string]*cobra.Command = make(map[string]*cobra.Command)

func GetComposeCmd(cmd string) *cobra.Command {
	_, cmds := WrappedCmd(ComposeFlags)
	return cmds[cmd]
}

// Handlers
func WrappedCmd(composeFlags models.ComposeFlags) ([]*cobra.Command, map[string]*cobra.Command) {
	// Docker stuff
	if plugin.RunningStandalone() && len(os.Args) > 2 && os.Args[1] == "compose" {
		os.Args = append([]string{"docker"}, compatibility.Convert(os.Args[2:])...)
	}
	// Create docker client
	op := &flags.ClientOptions{}
	if os.Getenv("DOCKER_CERT_PATH") != "" {
		TLSOptions := tlsconfig.Options{
			CAFile:   filepath.Join(os.Getenv("DOCKER_CERT_PATH"), "/ca.pem"),
			CertFile: filepath.Join(os.Getenv("DOCKER_CERT_PATH"), "/cert.pem"),
			KeyFile:  filepath.Join(os.Getenv("DOCKER_CERT_PATH"), "/key.pem"),
		}
		op = &flags.ClientOptions{
			TLSOptions: &TLSOptions,
		}
	}
	cliOptions := func(cli *command.DockerCli) error {
		cli.Initialize(op)
		return nil
	}
	dockerCli, err := command.NewDockerCli(cliOptions)
	if err != nil {
		debug.PrintStack()
		panic(err)
	}
	// Create docker command
	backend := compose.NewComposeService(dockerCli).(commands.Backend)
	cmdDocker := commands.RootCommand(dockerCli, backend)

	// Stuff to return
	cmds := []*cobra.Command{}
	mappedCmds := make(map[string]*cobra.Command)

	// Filter commands
	for _, c := range cmdDocker.Commands() {
		command := strings.Split(c.Use, " ")[0]
		if composeFlags.Contains(command) {
			// Filter flags
			flags := composeFlags[command].ExtractFlags(cmdDocker.Flags(), c.Flags())
			c.ResetFlags()
			c.Flags().AddFlagSet(flags)
			// Modify file flag
			if c.Flags().Lookup("file") != nil {
				modifyFileFlag(c)
			}
			// Save command
			cmds = append(cmds, c)
			mappedCmds[command] = c
		}
	}
	return cmds, mappedCmds
}

// Modify the file flag to be hidden and add a folder flag
func modifyFileFlag(c *cobra.Command) {
	// Modify flags
	c.Flags().Lookup("file").Hidden = true
	c.Flags().Lookup("file").Shorthand = ""
	stamusFlags.Config.AddAsFlag(c, false)
	// Save the command
	currentRunE := c.RunE
	// Modify cmd function
	c.RunE = makeCustomRunner(currentRunE)
}

// defaultRemoveOrphans turns on --remove-orphans for `up` unless the user set it
// explicitly. The rendered compose file is the source of truth, so services
// dropped from it (e.g. after `config set arkime=false`) are torn down on the
// next up rather than left running as untracked orphans.
func defaultRemoveOrphans(cmd *cobra.Command) {
	if cmd.Name() != "up" {
		return
	}
	if ro := cmd.Flags().Lookup("remove-orphans"); ro != nil &&
		!cmd.Flags().Changed("remove-orphans") {
		ro.Value.Set("true")
	}
}

// shouldCreateBackup returns true if the compose command should trigger a backup.
func shouldCreateBackup(cmdName string, volumesValue string) bool {
	return cmdName == "down" && volumesValue == "true"
}

// Return a custom runner for the command, that sets the file flag to the folder flag
func makeCustomRunner(
	runE func(cmd *cobra.Command, args []string) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Check if shutdown is in progress
		if shutdown.IsShuttingDown() {
			logging.Logger.Warn("Rejecting compose operation during shutdown",
				zap.String("command", cmd.Name()))
			return fmt.Errorf("operation rejected: shutdown in progress")
		}

		// Track this operation
		opID := fmt.Sprintf("compose-%s-%d", cmd.Name(), operationCounter.Add(1))
		_, done := shutdown.GetTracker().Start(shutdown.Context(), opID)
		defer done()

		// Get folder flag value
		configFlag := cmd.Flags().Lookup("config")
		conf := configFlag.Value.String()
		composeFile := utils.GetComposeFilePath(conf)

		// Create automatic backup before compose down --volumes
		volumesFlag := cmd.Flags().Lookup("volumes")
		volumesValue := ""
		if volumesFlag != nil {
			volumesValue = volumesFlag.Value.String()
		}
		if shouldCreateBackup(cmd.Name(), volumesValue) {
			// Extract config name from path
			configName := filepath.Base(conf)

			// Create backup
			_, err := backup.CreateBackup(configName, backup.BackupTypeAuto, logging.Logger)
			if err != nil {
				// Log warning but continue with operation
				logging.Logger.Warn("Failed to create backup before compose down --volumes",
					zap.String("config", configName),
					zap.Error(err),
				)
			}
		}

		// Set file flag
		fileFlag := cmd.Flags().Lookup("file")
		fileFlag.Value.Set(composeFile)
		fileFlag.DefValue = composeFile

		// Set project name flag from stored config
		// Resolve to absolute path since stamus config stores absolute paths
		absConf, _ := filepath.Abs(conf)
		projectName := stamus.GetProjectName(absConf)
		if projectName != "" {
			if projectFlag := cmd.Flags().Lookup("project-name"); projectFlag != nil {
				projectFlag.Value.Set(projectName)
			}
		}

		// For `up`, remove orphaned containers by default (opt out with
		// `--remove-orphans=false`).
		defaultRemoveOrphans(cmd)

		// Run existing command
		err := runE(cmd, args)
		if err != nil {
			logging.Sugar.Error(err)

			return err
		}
		return err
	}
}
