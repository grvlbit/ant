/*
Copyright © 2023 grvlbit

*/
package cmd

import (
	"os"
	"fmt"
	"github.com/spf13/cobra"
)



// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ant",
	Short: "An ansible toolkit",
	Long: `
ANT is a CLI for Ansible written in Go that empowers ansible role creation.
This application is a tool to generate the needed files
to quickly create a new ansible role from a template repository.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
            if len(args) == 0 {
	        err := cmd.Help()
	        if err != nil {
		    fmt.Printf("Error showing help: %v\n", err)
	            return
                }
                os.Exit(0)
            } 
	}, 
}


func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ant.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


