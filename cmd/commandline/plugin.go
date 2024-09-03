package main

import (
	init_pkg "github.com/langgenius/dify-plugin-daemon/cmd/commandline/init"
	"github.com/spf13/cobra"
)

var (
	pluginInitCommand = &cobra.Command{
		Use:   "init",
		Short: "Init",
		Long:  "Init",
		Run: func(c *cobra.Command, args []string) {
			init_pkg.InitPlugin()
		},
	}

	pluginModelCommand = &cobra.Command{
		Use:   "model",
		Short: "Model",
		Long:  "Model management for plugin",
	}

	pluginToolCommand = &cobra.Command{
		Use:   "tool",
		Short: "Tool",
		Long:  "Tool management for plugin",
	}

	pluginEndpointCommand = &cobra.Command{
		Use:   "endpoint",
		Short: "Endpoint",
		Long:  "Endpoint management for plugin",
	}

	pluginPackageCommand = &cobra.Command{
		Use:   "package",
		Short: "Package",
		Long:  "Package plugins",
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
		Use:   "add permission",
		Short: "",
		Long:  "Add permission to plugin, you can find the available permission by running `dify plugin permission`",
	}

	pluginPermissionDropCommand = &cobra.Command{
		Use:   "drop permission",
		Short: "",
		Long:  "Drop permission from plugin, you can find the available permission by running `dify plugin permission`",
	}
)

func init() {
	pluginCommand.AddCommand(pluginInitCommand)
	pluginCommand.AddCommand(pluginModelCommand)
	pluginCommand.AddCommand(pluginToolCommand)
	pluginCommand.AddCommand(pluginEndpointCommand)
	pluginCommand.AddCommand(pluginPackageCommand)
	pluginCommand.AddCommand(pluginPermissionCommand)
	pluginPermissionCommand.AddCommand(pluginPermissionAddCommand)
	pluginPermissionCommand.AddCommand(pluginPermissionDropCommand)
}
