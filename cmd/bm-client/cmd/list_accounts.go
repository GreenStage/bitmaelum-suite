package cmd

import (
	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-client/handlers"
	"github.com/bitmaelum/bitmaelum-suite/internal/vault"
	"github.com/spf13/cobra"
)

var listAccountsCmd = &cobra.Command{
	Use:     "list-accounts",
	Aliases: []string{"list-account", "ls", "list"},
	Short:   "List your accounts",
	Long:    `Displays a list of all your accounts currently available`,
	Run: func(cmd *cobra.Command, args []string) {
		v := vault.OpenVault()

		handlers.ListAccounts(v, *displayKeys)
	},
}

var displayKeys *bool

func init() {
	rootCmd.AddCommand(listAccountsCmd)

	displayKeys = listAccountsCmd.Flags().BoolP("keys", "k", false, "Display private and public key")
}
