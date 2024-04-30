package utils

import (
	// Common
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	// External
	"github.com/Masterminds/semver/v3"

	// Custom
	"stamus-ctl/internal/logging"
)

func GetExecExist(executable string, flags ...string) bool {
	// Format cmd
	flags = append([]string{executable}, flags...)
	cmd := exec.Command("which", flags...)

	// Execute command which
	output, err := cmd.Output()

	// If the command fails or output is empty, then executable not found
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		logging.Sugar.Errorw("Executable '%s' not found\n", "executable", executable)
		return false
	}

	// Else executable was found
	fmt.Printf("Executable '%s' found at:\n%s\n", executable, output)
	return true
}

func GetExecVersion(executable string, flags ...string) (*semver.Version, error) {
	// Format cmd
	flags = append([]string{"version"}, flags...)
	cmd := exec.Command(executable, flags...)

	// Give output pointer
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	//Run cmd
	if err := cmd.Run(); err != nil {
		logging.Sugar.Errorw("cannot fetch version.", "error", err, "exec", executable)
		return nil, fmt.Errorf("cannot %s fetch version", executable)
	}

	//Format output
	output := stdout.String()
	splited := strings.Split(output, " ")
	extracted := strings.Trim(splited[len(splited)-1], "\n")
	version, err := semver.NewVersion(extracted)
	if err != nil {
		logging.Sugar.Errorw("cannot parse version.", "error", err, "exec", extracted)
		return nil, fmt.Errorf("cannot parse %s version", executable)
	}

	// Return
	logging.Sugar.Debugw("detected version.", "version", version, "executable", executable)
	return version, nil
}
