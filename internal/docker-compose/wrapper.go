package compose

import (
	// Core
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	// Common
	stamusFlags "stamus-ctl/internal/handlers"
	"stamus-ctl/internal/models"

	// External
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/cmd/compatibility"
	commands "github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/spf13/cobra"
)

// Constants
var ComposeFlags = models.ComposeFlags{
	"up": models.CreateComposeFlags(
		[]string{"file"},
		[]string{"detach"},
	),
	"down": models.CreateComposeFlags(
		[]string{"file"},
		[]string{"volumes", "remove-orphans"},
	),
	"ps": models.CreateComposeFlags(
		[]string{"file"},
		[]string{"services", "quiet", "format"},
	),
	"logs": models.CreateComposeFlags(
		[]string{"file"},
		[]string{"timestamps", "tail", "since", "until"},
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
	stamusFlags.Config.AddAsFlag(c, false)
	// Save the command
	currentRunE := c.RunE
	// Modify cmd function
	c.RunE = makeCustomRunner(currentRunE)
}

// Return a custom runner for the command, that sets the file flag to the folder flag
func makeCustomRunner(
	runE func(cmd *cobra.Command, args []string) error,
) func(cmd *cobra.Command, args []string) error {

	return func(cmd *cobra.Command, args []string) error {
		// Get folder flag value
		configFlag := cmd.Flags().Lookup("config")
		conf := configFlag.Value.String()
		possibleComposeFiles := []string{
			"docker-compose.yaml",
			"docker-compose.yml",
			"compose.yaml",
			"compose.yml",
		}
		composeFile := ""
		for _, file := range possibleComposeFiles {
			filePath := filepath.Join(conf, file)
			if _, err := os.Stat(filePath); err == nil {
				composeFile = filePath
				break
			}
		}
		// Set file flag
		fileFlag := cmd.Flags().Lookup("file")
		fileFlag.Value.Set(composeFile)
		fileFlag.DefValue = composeFile
		// Run existing command
		return runE(cmd, args)
	}
}
