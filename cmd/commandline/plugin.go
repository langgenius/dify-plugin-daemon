package main

import (
	"fmt"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/cmd/commandline/plugin"
	"github.com/spf13/cobra"
)

var (
	author                   string
	name                     string
	repo                     string
	description              string
	allowRegisterEndpoint    bool
	allowInvokeTool          bool
	allowInvokeModel         bool
	allowInvokeLLM           bool
	allowInvokeTextEmbedding bool
	allowInvokeRerank        bool
	allowInvokeTTS           bool
	allowInvokeSpeech2Text   bool
	allowInvokeModeration    bool
	allowInvokeNode          bool
	allowInvokeApp           bool
	allowUseStorage          bool
	storageSize              uint64
	category                 string
	language                 string
	minDifyVersion           string
	quick                    bool
	maxSizeMB                int64
	pluginInitCommand        = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new plugin",
		Long: `Initialize a new plugin with the given parameters.
If no parameters are provided, an interactive mode will be started.`,
		Run: func(c *cobra.Command, args []string) {
			plugin.InitPluginWithFlags(
				author,
				name,
				repo,
				description,
				allowRegisterEndpoint,
				allowInvokeTool,
				allowInvokeModel,
				allowInvokeLLM,
				allowInvokeTextEmbedding,
				allowInvokeRerank,
				allowInvokeTTS,
				allowInvokeSpeech2Text,
				allowInvokeModeration,
				allowInvokeNode,
				allowInvokeApp,
				allowUseStorage,
				storageSize,
				category,
				language,
				minDifyVersion,
				quick,
			)
		},
	}

	pluginEditPermissionCommand = &cobra.Command{
		Use:   "permission [plugin_path]",
		Short: "Edit permission",
		Long:  "Edit permission",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			plugin.EditPermission(args[0])
		},
	}

	pluginPackageCommand = &cobra.Command{
		Use:   "package [plugin_path]",
		Short: "Package",
		Long:  "Package plugins",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			inputPath := filepath.Clean(args[0])

			// using filename of input_path as output_path if not specified
			outputPath := ""

			if cmd.Flag("output_path").Value.String() != "" {
				outputPath = cmd.Flag("output_path").Value.String()
			} else {
				base := filepath.Base(inputPath)
				if base == "." || base == "/" {
					fmt.Println("Error: invalid input path, you should specify the path outside of plugin directory")
					return
				}
				outputPath = base + ".difypkg"
			}

			// read max-size flag (in MB), default 50
			maxSizeMB, err := cmd.Flags().GetInt64("max-size")
			if err != nil {
				maxSizeMB = 50
			}
			if maxSizeMB <= 0 {
				maxSizeMB = 50
			}
			maxSizeBytes := maxSizeMB * 1024 * 1024
			plugin.PackagePlugin(inputPath, outputPath, maxSizeBytes)
		},
	}

	pluginChecksumCommand = &cobra.Command{
		Use:   "checksum [plugin_path]",
		Short: "Checksum",
		Long:  "Calculate the checksum of the plugin, you need specify the plugin path or .difypkg file path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pluginPath := args[0]
			plugin.CalculateChecksum(pluginPath)
		},
	}

	pluginModuleCommand = &cobra.Command{
		Use:   "module",
		Short: "Module",
		Long:  "Module",
	}

	pluginModuleListCommand = &cobra.Command{
		Use:   "list [plugin_path]",
		Short: "List",
		Long:  "List modules",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pluginPath := args[0]
			plugin.ModuleList(pluginPath)
		},
	}

	pluginModuleAppendCommand = &cobra.Command{
		Use:   "append",
		Short: "Append",
		Long:  "Append",
	}

	pluginModuleAppendToolsCommand = &cobra.Command{
		Use:   "tools [plugin_path]",
		Short: "Tools",
		Long:  "Append tools",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pluginPath := args[0]
			plugin.ModuleAppendTools(pluginPath)
		},
	}

	pluginModuleAppendEndpointsCommand = &cobra.Command{
		Use:   "endpoints [plugin_path]",
		Short: "Endpoints",
		Long:  "Append endpoints",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pluginPath := args[0]
			plugin.ModuleAppendEndpoints(pluginPath)
		},
	}

	pluginReadmeCommand = &cobra.Command{
		Use:   "readme",
		Short: "Readme",
		Long:  "Readme",
	}

	pluginReadmeListCommand = &cobra.Command{
		Use:   "list [plugin_path]",
		Short: "List available README languages",
		Long:  "List available README languages in the specified plugin",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pluginPath := args[0]
			plugin.ListReadme(pluginPath)
		},
	}

	// NOTE: tester is deprecated, maybe, in several months, we will support this again
	// pluginTestCommand = &cobra.Command{
	// 	Use:   "test [-i inputs] [-t timeout] package_path invoke_type invoke_action",
	// 	Short: "",
	// 	Long: "Test runs the given plugin package locally, and you can specify the inputs using json format, if not specified, will use default inputs\n" +
	// 		"type: invoke type, available values: \n" +
	// 		"[\n" +
	// 		"	tool, model, endpoint\n" +
	// 		"]\n" +
	// 		"action: invoke action, available values: \n" +
	// 		"[\n" +
	// 		"	invoke_tool, validate_tool_credentials, \n" +
	// 		"	invoke_endpoint\n" +
	// 		"	invoke_llm, invoke_text_embedding, invoke_rerank, invoke_tts, invoke_speech2text, invoke_moderation, \n" +
	// 		"	validate_provider_credentials, validate_model_credentials, get_tts_model_voices, \n" +
	// 		"	get_text_embedding_num_tokens, get_ai_model_schemas, get_llm_num_tokens\n" +
	// 		"]\n",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		if len(args) < 3 {
	// 			log.Error("invalid args, please specify package_path, invoke_type, invoke_action")
	// 			return
	// 		}
	// 		// get package path
	// 		package_path_str := args[0]
	// 		// get invoke type
	// 		invoke_type_str := args[1]
	// 		// get invoke action
	// 		invoke_action_str := args[2]
	// 		// get inputs if specified
	// 		inputs := map[string]any{}
	// 		if cmd.Flag("inputs") != nil {
	// 			inputs_str := cmd.Flag("inputs").Value.String()
	// 			err := json.Unmarshal([]byte(inputs_str), &inputs)
	// 			if err != nil {
	// 				log.Error("failed to unmarshal inputs, inputs: %s, error: %v", inputs_str, err)
	// 				return
	// 			}
	// 		}
	// 		// parse flag
	// 		timeout := ""
	// 		if cmd.Flag("timeout") != nil {
	// 			timeout = cmd.Flag("timeout").Value.String()
	// 		}

)

func init() {
	pluginCommand.AddCommand(pluginInitCommand)
	pluginCommand.AddCommand(pluginPackageCommand)
	pluginCommand.AddCommand(pluginChecksumCommand)
	pluginCommand.AddCommand(pluginEditPermissionCommand)
	pluginCommand.AddCommand(pluginModuleCommand)
	pluginCommand.AddCommand(pluginReadmeCommand)
	pluginModuleCommand.AddCommand(pluginModuleListCommand)
	pluginModuleCommand.AddCommand(pluginModuleAppendCommand)
	pluginModuleAppendCommand.AddCommand(pluginModuleAppendToolsCommand)
	pluginModuleAppendCommand.AddCommand(pluginModuleAppendEndpointsCommand)
	pluginReadmeCommand.AddCommand(pluginReadmeListCommand)

	pluginInitCommand.Flags().StringVar(&author, "author", "", "Author name (1-64 characters, lowercase letters, numbers, dashes and underscores only)")
	pluginInitCommand.Flags().StringVar(&name, "name", "", "Plugin name (1-128 characters, lowercase letters, numbers, dashes and underscores only)")
	pluginInitCommand.Flags().StringVar(&description, "description", "", "Plugin description (cannot be empty)")
	pluginInitCommand.Flags().StringVar(&repo, "repo", "", "Plugin repository URL (optional)")
	pluginInitCommand.Flags().BoolVar(&allowRegisterEndpoint, "allow-endpoint", false, "Allow the plugin to register endpoints")
	pluginInitCommand.Flags().BoolVar(&allowInvokeTool, "allow-tool", false, "Allow the plugin to invoke tools")
	pluginInitCommand.Flags().BoolVar(&allowInvokeModel, "allow-model", false, "Allow the plugin to invoke models")
	pluginInitCommand.Flags().BoolVar(&allowInvokeLLM, "allow-llm", false, "Allow the plugin to invoke LLM models")
	pluginInitCommand.Flags().BoolVar(&allowInvokeTextEmbedding, "allow-text-embedding", false, "Allow the plugin to invoke text embedding models")
	pluginInitCommand.Flags().BoolVar(&allowInvokeRerank, "allow-rerank", false, "Allow the plugin to invoke rerank models")
	pluginInitCommand.Flags().BoolVar(&allowInvokeTTS, "allow-tts", false, "Allow the plugin to invoke TTS models")
	pluginInitCommand.Flags().BoolVar(&allowInvokeSpeech2Text, "allow-speech2text", false, "Allow the plugin to invoke speech to text models")
	pluginInitCommand.Flags().BoolVar(&allowInvokeModeration, "allow-moderation", false, "Allow the plugin to invoke moderation models")
	pluginInitCommand.Flags().BoolVar(&allowInvokeNode, "allow-node", false, "Allow the plugin to invoke nodes")
	pluginInitCommand.Flags().BoolVar(&allowInvokeApp, "allow-app", false, "Allow the plugin to invoke apps")
	pluginInitCommand.Flags().BoolVar(&allowUseStorage, "allow-storage", false, "Allow the plugin to use storage")
	pluginInitCommand.Flags().Uint64Var(&storageSize, "storage-size", 0, "Maximum storage size in bytes")
	pluginInitCommand.Flags().StringVar(&category, "category", "", `Plugin category. Available options:
  - tool: Tool plugin
  - llm: Large Language Model plugin
  - text-embedding: Text embedding plugin
  - speech2text: Speech to text plugin
  - moderation: Content moderation plugin
  - rerank: Rerank plugin
  - tts: Text to speech plugin
  - extension: Extension plugin
  - agent-strategy: Agent strategy plugin`)
	pluginInitCommand.Flags().StringVar(&language, "language", "", `Programming language. Available options:
  - python: Python language`)
	pluginInitCommand.Flags().StringVar(&minDifyVersion, "min-dify-version", "", "Minimum Dify version required")
	pluginInitCommand.Flags().BoolVar(&quick, "quick", false, "Skip interactive mode and create plugin directly")

	pluginPackageCommand.Flags().StringP("output_path", "o", "", "output path")
	pluginPackageCommand.Flags().Int64Var(&maxSizeMB, "max-size", 50, "Maximum uncompressed size in MB")
}
