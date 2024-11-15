package main

import (
	"fmt"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/cmd/commandline/plugin"
	"github.com/spf13/cobra"
)

var (
	pluginInitCommand = &cobra.Command{
		Use:   "init",
		Short: "Init",
		Long:  "Init",
		Run: func(c *cobra.Command, args []string) {
			plugin.InitPlugin()
		},
	}

	pluginPackageCommand = &cobra.Command{
		Use:   "package plugin_path [-o output_path]",
		Short: "Package",
		Long:  "Package plugins",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			inputPath := args[0]

			// using filename of input_path as output_path if not specified
			outputPath := ""

			if cmd.Flag("output_path") != nil {
				outputPath = cmd.Flag("output_path").Value.String()
			} else {
				base := filepath.Base(inputPath)
				if base == "." || base == "/" {
					fmt.Println("Error: invalid input path, you should specify the path outside of plugin directory")
					return
				}
				outputPath = base + ".difypkg"
			}

			plugin.PackagePlugin(inputPath, outputPath)
		},
	}

	pluginChecksumCommand = &cobra.Command{
		Use:   "checksum plugin_path",
		Short: "Checksum",
		Long:  "Calculate the checksum of the plugin, you need specify the plugin path or .difypkg file path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pluginPath := args[0]
			plugin.CalculateChecksum(pluginPath)
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
	// pluginCommand.AddCommand(pluginTestCommand)
	// pluginTestCommand.Flags().StringP("inputs", "i", "", "inputs")
	// pluginTestCommand.Flags().StringP("timeout", "t", "", "timeout")
}
