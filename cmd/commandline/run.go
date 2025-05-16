package main

import (
	"github.com/langgenius/dify-plugin-daemon/cmd/commandline/run"
	"github.com/spf13/cobra"
)

/*
 Test is a very important component of Dify plugins, to ensure every plugin is working as expected
 We must provide a way to make the test a pipeline and use standard CI/CD tools to run the tests

 However, developers prefer to write test codes in language like Python, it's hard to enforce them to use Go
 and what we need is actually a way to launch plugins locally

 It makes things easier, the command should be `run`, instead of `test`, user could use `dify plugin run <plugin_id>`
 to launch and test it through stdin/stdout
*/

var (
	runPluginPayload run.RunPluginPayload
)

var (
	runPluginCommand = &cobra.Command{
		Use:   "run [plugin_package_path]",
		Short: "run",
		Long:  "Launch a plugin locally and communicate through stdin/stdout or TCP",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			runPluginPayload.PluginPath = args[0]
			// launch plugin
			run.RunPlugin(runPluginPayload)
		},
	}
)

func init() {
	pluginCommand.AddCommand(runPluginCommand)

	runPluginCommand.Flags().StringVarP(&runPluginPayload.RunMode, "mode", "m", "stdio", "run mode, stdio or tcp")
	runPluginCommand.Flags().BoolVarP(&runPluginPayload.EnableLogs, "enable-logs", "l", false, "enable logs")
	runPluginCommand.Flags().StringVarP(&runPluginPayload.ResponseFormat, "response-format", "r", "text", "response format, text or json")
}
