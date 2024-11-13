package plugin

func TestPlugin(pluginPath string) {
	// 		// get invoke_type and invoke_action
	// 		invoke_type := access_types.PluginAccessType(invoke_type_str)
	// 		if !invoke_type.IsValid() {
	// 			log.Error("invalid invoke type: %s", invoke_type_str)
	// 			return
	// 		}
	// 		invoke_action := access_types.PluginAccessAction(invoke_action_str)
	// 		if !invoke_action.IsValid() {
	// 			log.Error("invalid invoke action: %s", invoke_action_str)
	// 			return
	// 		}

	// 		// init routine pool
	// 		routine.InitPool(1024)

	// 		// clean working directory when test finished
	// 		defer os.RemoveAll("./working")

	// 		// init testing config
	// 		config := &app.Config{
	// 			PluginWorkingPath:    "./working/cwd",
	// 			PluginStoragePath:    "./working/storage",
	// 			PluginMediaCachePath: "./working/media_cache",
	// 			ProcessCachingPath:   "./working/subprocesses",
	// 			Platform:             app.PLATFORM_LOCAL,
	// 		}
	// 		config.SetDefault()

	// 		// init oss
	// 		oss := local.NewLocalStorage("./storage")

	// 		// init plugin manager
	// 		plugin_manager := plugin_manager.InitGlobalManager(oss, config)

	// 		response, err := plugin_manager.TestPlugin(package_path_str, inputs, invoke_type, invoke_action, timeout)
	// 		if err != nil {
	// 			log.Error("failed to test plugin, package_path: %s, error: %v", package_path_str, err)
	// 			return
	// 		}

	// 		for response.Next() {
	// 			item, err := response.Read()
	// 			if err != nil {
	// 				log.Error("failed to read response item, error: %v", err)
	// 				return
	// 			}
	// 			log.Info("%v", parser.MarshalJson(item))
	// 		}
	// 	},
	// }
}
