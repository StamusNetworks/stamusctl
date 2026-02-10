package completion

import (
	"os"

	"github.com/spf13/cobra"
)

// CompletionCmd returns the completion command that generates shell completion scripts
func CompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for stamusctl.

To load completions:

Bash:
  # Add this to your ~/.bashrc or run it directly:
  $ source <(stamusctl completion bash)

  # Or generate a file and load it:
  $ stamusctl completion bash > /etc/bash_completion.d/stamusctl

Zsh:
  # If shell completion is not already enabled, enable it:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # Add this to your ~/.zshrc or run it directly:
  $ source <(stamusctl completion zsh)

  # Or install to fpath for persistent completion:
  $ stamusctl completion zsh > "${fpath[1]}/_stamusctl"

Fish:
  $ stamusctl completion fish | source

  # Or install for persistent completion:
  $ stamusctl completion fish > ~/.config/fish/completions/stamusctl.fish

PowerShell:
  PS> stamusctl completion powershell | Out-String | Invoke-Expression

  # Or add to your PowerShell profile for persistent completion:
  PS> stamusctl completion powershell >> $PROFILE
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
