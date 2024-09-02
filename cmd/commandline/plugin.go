package main

import (
	"github.com/langgenius/dify-plugin-daemon/cmd/commandline/cmd"
	"github.com/spf13/cobra"
)

var (
	pluginInitCommand = &cobra.Command{
		Use:   "init",
		Short: "Init",
		Long:  "Init",
		Run: func(c *cobra.Command, args []string) {
			cmd.InitPlugin()
		},
	}

	pluginPermissionCommand = &cobra.Command{
		Use:   "permission",
		Short: "Permission",
		Long: `Permission, available values: 
tools					- allow plugin to call tools
models					- allow plugin to call models
models.llm				- allow plugin to call llm
models.text_embedding			- allow plugin to call text_embedding model
models.rerank				- allow plugin to call rerank model
models.tts				- allow plugin to call tts
models.speech2text			- allow plugin to call speech2text
models.moderation			- allow plugin to call moderation
apps					- allow plugin to call apps
storage					- allow plugin to use storage
endpoint				- allow plugin to register endpoint`,
	}

	pluginPermissionAddCommand = &cobra.Command{
		Use:   "add",
		Short: "Add permission to plugin",
		Long:  "Add permission to plugin, you can find the available permission by running `dify plugin permission`",
	}

	pluginPermissionDropCommand = &cobra.Command{
		Use:   "drop",
		Short: "Drop permission from plugin",
		Long:  "Drop permission from plugin, you can find the available permission by running `dify plugin permission`",
	}
)

func init() {
	pluginCommand.AddCommand(pluginInitCommand)
	pluginCommand.AddCommand(pluginPermissionCommand)
	pluginPermissionCommand.AddCommand(pluginPermissionAddCommand)
	pluginPermissionCommand.AddCommand(pluginPermissionDropCommand)
}
