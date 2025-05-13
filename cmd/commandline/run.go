//go:build fullfeatures

package main

import "github.com/spf13/cobra"

/*
 Test is a very important component of Dify plugins, to ensure every plugin is working as expected
 We must provide a way to make the test a pipeline and use standard CI/CD tools to run the tests

 However, developers prefer to write test codes in language like Python, it's hard to enforce them to use Go
 and what we need is actually a way to launch plugins locally

 It makes things easier, the command should be `run`, instead of `test`, user could use `dify plugin run <plugin_id>`
 to launch and test it through stdin/stdout
*/

var (
	runPluginCommand = &cobra.Command{
		Use:   "run",
		Short: "run",
		Long:  "Launch a plugin locally and communicate through stdin/stdout",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			// launch plugin
		},
	}
)

func init() {
	rootCommand.AddCommand(runPluginCommand)
}
