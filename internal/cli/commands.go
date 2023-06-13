package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dezier [options]",
	Short: "dezier is DesmosBezierRenderer tool",
}

func Run() {
	if err := rootCmd.Execute(); err != nil {
		errors.Wrap(err, "failed to execute root command")
	}
}
