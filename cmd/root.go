package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"mcap-utility/cmd/edit"
	"mcap-utility/cmd/info"
	"mcap-utility/internal/constants"
)

var (
	rootCmd = &cobra.Command{
		Use:   "mcap-utility",
		Short: "A CLI for manipulating MCAP files.",
		Long: fmt.Sprintf(
			`mcap-utility is a command line interface for (%s) file manipulations.
Authored by Thomas Pham
`, constants.MCAPFIleExtension,
		),
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(info.InfoCmd)
	rootCmd.AddCommand(edit.EditCmd)
}
