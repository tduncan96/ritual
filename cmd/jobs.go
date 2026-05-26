package cmd

import (
	"fmt"

	"ritual/internal/io"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import jobs from external source",
}

var importTomlCmd = &cobra.Command{
	Use:   "toml <file>",
	Short: ".toml I/O for Jobs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := io.TomlToJob(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Imported Job ID: %d\n", id)
		return nil
	},
}

func init() {
	importCmd.AddCommand(importTomlCmd)
	rootCmd.AddCommand(importCmd)
}
