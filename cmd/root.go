package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nap",
	Short: "This is a tool to install nix packages the easy way.",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
