package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func testCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run configuration tests before applying",
		Long: `Run test scripts from the configuration's tests/ directory.

Discovers and executes all .sh and .nix test files in <config>/tests/,
reporting pass/fail for each. Shell scripts are run with bash; Nix files
are evaluated with nix-instantiate --eval.

Exit code is non-zero if any test fails.

Examples:
  # Run all tests for the current configuration
  stamusctl nix test

  # Run tests for a specific configuration
  stamusctl nix test --config /path/to/my-config

  # Run only tests matching a pattern
  stamusctl nix test --filter "check-*.sh"
`,
		RunE:         testHandler,
		SilenceUsage: true,
	}

	flags.Config.AddAsFlag(cmd, false)
	cmd.Flags().StringP("filter", "f", "", "Glob pattern to filter test files (e.g. \"check-*.sh\")")

	return cmd
}

func testHandler(cmd *cobra.Command, _ []string) error {
	conf, err := flags.Config.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	filter, _ := cmd.Flags().GetString("filter")

	return handlers.NixTestHandler(handlers.NixTestHandlerInputs{
		Config: conf.(string),
		Filter: filter,
	})
}
