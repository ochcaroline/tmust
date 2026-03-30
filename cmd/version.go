package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

// SetVersion is called from main to inject the build-time version string.
func SetVersion(v string) {
	version = v
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
