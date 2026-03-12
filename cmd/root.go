package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-boilerplate",
	Short: "Generate production-ready Go backend projects",
	Long: `Interactive CLI generator for Go backends with clean architecture,
dual protocol support (REST + gRPC), and optional feature toggles.

Examples:
  go-boilerplate init                    # Interactive mode
  go-boilerplate init --name=my-api \
    --module=github.com/org/my-api \
    --features=auth,observability,docker`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
