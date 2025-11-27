package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nostar",
		Short: "Nostr simple relay server",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "not yet implemented")
			return nil
		},
	}

	cmd.AddCommand()

	return cmd
}

func main() {
	rootCmd := newRootCommand()
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
