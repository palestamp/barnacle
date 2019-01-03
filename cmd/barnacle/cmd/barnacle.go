package cmd

import (
	"github.com/spf13/cobra"
)

// New returns an initialized command tree.
func New() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "barnacle",
		Short: "Barnacle is a queue management agent.",
	}

	rootCmd.AddCommand(ServeCmd())
	return rootCmd
}
