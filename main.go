package main

import (
	"fmt"
	"os"

	"github.com/finos/common-cloud-controls/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ccc",
	Short: "CCC — Common Cloud Controls CLI",
}

func init() {
	rootCmd.AddCommand(cmd.GenerateCmd)
	rootCmd.AddCommand(cmd.ReleaseCmd)
	rootCmd.AddCommand(cmd.PublishCmd)
	rootCmd.AddCommand(cmd.PreviewLocalCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
