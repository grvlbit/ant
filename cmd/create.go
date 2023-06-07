/*
Copyright © 2023 grvlbit
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "create",
	Short: "create ansible role from template",
	Long: `
Initialize a new ansible role from a template repository. 
Different platforms for github actions and molecule testing 
are supported and automatically configured.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating a new role from template. Let's get started!")
		createRole()
	},
}

// the questions to ask
var qs = []*survey.Question{
	{
		Name:     "name",
		Prompt:   &survey.Input{Message: "What is your role called?"},
		Validate: survey.Required,
	},
	{
		Name:     "author",
		Prompt:   &survey.Input{Message: "What is your name?"},
		Validate: survey.Required,
	},
	{
		Name:     "company",
		Prompt:   &survey.Input{Message: "What is your company?"},
		Validate: survey.Required,
	},
	{
		Name:     "namespace",
		Prompt:   &survey.Input{Message: "Which namespace controls the role?"},
		Validate: survey.Required,
	},
	{
		Name: "platforms",
		Prompt: &survey.MultiSelect{
			Message: "Choose one or more platform:",
			Options: []string{
				"ubuntu2004",
				"ubuntu2204",
				"rockylinux8",
				"rockylinux9",
			},
		},
		Validate: survey.Required,
	},
	{
		Name: "license",
		Prompt: &survey.Select{
			Message: "Choose a license:",
			Options: []string{"GPL-2.0-or-later", "GPL-3.0-or-later", "MIT", "BSD"},
			Default: "MIT",
		},
	},
	{
		Name:     "description",
		Prompt:   &survey.Input{Message: "Please enter a short role description."},
		Validate: survey.Required,
	},
	{
		Name:     "gitinit",
		Prompt:   &survey.Confirm{Message: "Do you want to init the role as git repository?"},
		Validate: survey.Required,
	},
}

type Metadata struct {
	Name        string
	Description string
	Company     string
	License     string
	Author      string
	Namespace   string
	Platforms   []string
	Gitinit     bool
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// confirm function asks for user input
// returns bool
func confirm() bool {

	var input string

	fmt.Printf("Do you want to continue with this operation? [y|n]: ")
	_, err := fmt.Scanln(&input)
	if err != nil {
		panic(err)
	}
	input = strings.ToLower(input)

	if input == "y" || input == "yes" {
		return true
	}
	return false

}

func cleanup(dir string, err *error) {
	// Remove the cloned repository
	e := os.RemoveAll(dir)
	switch *err {
	case nil:
		*err = e
	default:
		if e != nil {
			log.Println("Cleanup failed:", e)
		}
	}
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createRole() {

	// the answers will be written to this struct
	meta := Metadata{}

	// perform the questions
	err := survey.Ask(qs, &meta)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Please check you inputs. Is everything correct?")
	if !confirm() {
		fmt.Println("Aborting")
		return
	}

	// Destination directory for copying files (excluding .git)
	copyDir := meta.Namespace + "." + meta.Name

	if _, err := os.Stat(copyDir); !os.IsNotExist(err) {
		fmt.Printf("Error creating role directory: Directory exists.\n")
		return
	}

	// URL of the repository to clone
	repoURL := "https://github.com/hpc-unibe-ch/ansible-role-template.git"

	// Destination directory for cloning the repository
	destDir, err := os.MkdirTemp("", "ant-")
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err)
		return
	}
	defer cleanup(destDir, &err)

	// Clone the repository to the destination directory
	_, err = git.PlainClone(destDir, false, &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", "ant")),
		SingleBranch:  true,
	})
	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		return
	}

	// Regular expression pattern to match and replace
	pattern := "template"
	replace := meta.Name

	// Compile the regular expression pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("Error compiling regular expression: %v\n", err)
		return
	}

	// Create the copy directory if it doesn't exist
	if meta.Gitinit {
		fmt.Printf("Initializing new git repo: %v\n", meta.Gitinit)
		cmd := exec.Command("git", "init", "-b", "main", copyDir)
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error initializing new repository: %v\n", err)
			return
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("Error initializing new repository: %v\n", err)
			return
		}
	} else {
		err = os.MkdirAll(copyDir, 0755)
		if err != nil {
			fmt.Printf("Error creating copy directory: %v\n", err)
			return
		}
	}

	// Walk through the cloned repository files and copy them to the copy directory
	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" {
				// Skip the .git directory
				return filepath.SkipDir
			}

			// Create the corresponding directory in the copy directory
			relPath, err := filepath.Rel(destDir, path)
			if err != nil {
				return err
			}
			copyPath := filepath.Join(copyDir, relPath)
			err = os.MkdirAll(copyPath, 0755)
			if err != nil {
				return err
			}
		} else {
			if filepath.Base(path) != ".git" {
				// Read the file contents
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				// Apply the regular expression replacement
				data = re.ReplaceAll(data, []byte(replace))

				// Apply go templates
				tmpl, err := template.New("file").Delims("<<", ">>").Parse(string(data))
				if err != nil {
					return err
				}

				//var buf strings.Builder
				var buf bytes.Buffer
				err = tmpl.Execute(&buf, meta)
				if err != nil {
					return err
				}

				// Create the corresponding file in the copy directory
				relPath, err := filepath.Rel(destDir, path)
				if err != nil {
					return err
				}
				copyPath := filepath.Join(copyDir, relPath)
				err = os.WriteFile(copyPath, buf.Bytes(), 0644)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error copying files: %v\n", err)
		return
	}

	fmt.Println("Repository cloned, files modified, and copied successfully.")

}
