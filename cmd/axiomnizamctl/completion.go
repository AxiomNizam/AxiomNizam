package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for bash, zsh, fish, or powershell.

Install bash completion:
  axiomnizamctl completion bash | sudo tee /etc/bash_completion.d/axiomnizamctl
  
Install zsh completion:
  axiomnizamctl completion zsh | sudo tee /usr/local/share/zsh/site-functions/_axiomnizamctl
  
Install fish completion:
  axiomnizamctl completion fish | sudo tee /usr/share/fish/vendor_completions.d/axiomnizamctl.fish
  
Install powershell completion:
  axiomnizamctl completion powershell | Out-String | Out-File -FilePath $PROFILE.CurrentUserCurrentHost
`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			RootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			RootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			RootCmd.GenPowerShellCompletion(os.Stdout)
		}
	},
}
