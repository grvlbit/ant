/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "wrapper to yamllint and ansible-lint",
	Long: `This command will perform ansible linting on 
the current directory. By default yamllint and ansible-lint
are used. If the commands are not availble the respective tool
will be skipped.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Linting. Let's get started!")
		lint()
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lintCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lintCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func registerLinters() string {
	linters := ""
	if commandExists("yamllint") {
		linters = linters + "yamllint "
	}
	if commandExists("ansible-lint") {
		linters = linters + "ansible-lint"
	}
	return linters
}

func lint() {
	fmt.Printf("Checking available linters..\n")
	linters := registerLinters()
	if linters != "" {
		fmt.Printf("Found %s\n", linters)
	} else {
		fmt.Printf("No linters found. Please consider installing yamllint and ansible-lint")
		return
	}

	if commandExists("yamllint") {
		out, err := exec.Command("yamllint", ".").Output()
		if err != nil {
			fmt.Printf("Error linting directory with yamllint: %v\n", err)
			log.Fatal(err)
		}
		fmt.Printf("%s\n", out)
	} else {
		fmt.Printf("yamllint not found in $PATH. Please consider installing it to use ant lint.")
	}

	if commandExists("ansible-lint") {
		out, err := exec.Command("ansible-lint").Output()
		if err != nil {
			fmt.Printf("Error linting directory with ansible-lint: %v\n", err)
			log.Fatal(err)
		}
		fmt.Printf("%s\n", out)
	} else {
		fmt.Printf("ansible-lint not found in $PATH. Please consider installing it to use ant lint.")
	}
}
