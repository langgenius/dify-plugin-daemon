package main

import (
	"github.com/langgenius/dify-plugin-daemon/cmd/commandline/model"
	"github.com/spf13/cobra"
)

var (
	modelTemplatesCommand = &cobra.Command{
		Use:   "templates [-t provider|model] [-m model_type] [name]",
		Short: "Templates",
		Long:  "List all model templates, you can use it to create new model",
		Run: func(cmd *cobra.Command, args []string) {
			// get provider or model
			typ, _ := cmd.Flags().GetString("type")
			// get model_type
			model_type, _ := cmd.Flags().GetString("model_type")
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			model.ListTemplates(typ, model_type, name)
		},
	}

	newProviderCommand = &cobra.Command{
		Use:   "provider [template] name",
		Short: "Provider",
		Long:  "Using template to create new provider, one plugin only support one provider",
	}

	newModelCommand = &cobra.Command{
		Use:   "new [template] name",
		Short: "Model",
		Long:  "Using template to create new model, you need to create a provider first",
	}
)

func init() {
	pluginModelCommand.AddCommand(modelTemplatesCommand)
	pluginModelCommand.AddCommand(newProviderCommand)
	pluginModelCommand.AddCommand(newModelCommand)
}
