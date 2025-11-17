package debugging_runtime

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/service/debugging_service"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func (d *DifyServer) handleHandleShake(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) (*debugging_service.ConnectionInfo, error) {
	if runtime.handshake {
		return nil, errors.New("handshake already completed")
	}

	key, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterHandshake](registerPayload.Data)
	if err != nil {
		// close connection if handshake failed
		return nil, errors.New("handshake failed, invalid handshake message")
	}

	info, err := debugging_service.GetConnectionInfo(key.Key)
	if err == cache.ErrNotFound {
		// close connection if handshake failed
		return nil, errors.New("handshake failed, invalid key")
	} else if err != nil {
		// close connection if handshake failed
		return nil, fmt.Errorf("failed to get connection info: %v", err)
	}

	return info, nil
}

func (d *DifyServer) handleAssetsTransfer(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	assetChunk, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterAssetChunk](registerPayload.Data)
	if err != nil {
		return fmt.Errorf("transfer assets failed, error: %v", err)
	}

	buffer, ok := runtime.assets[assetChunk.Filename]
	if !ok {
		runtime.assets[assetChunk.Filename] = &bytes.Buffer{}
		buffer = runtime.assets[assetChunk.Filename]
	}

	// allows at most 50MB assets
	if runtime.assetsBytes+int64(len(assetChunk.Data)) > 50*1024*1024 {
		return errors.New("assets too large, at most 50MB")
	}

	// decode as base64
	data, err := base64.StdEncoding.DecodeString(assetChunk.Data)
	if err != nil {
		return fmt.Errorf("assets decode failed, error: %v", err)
	}

	buffer.Write(data)

	// update assets bytes
	runtime.assetsBytes += int64(len(data))

	return nil
}

func (d *DifyServer) handleInitializationEndEvent(
	runtime *RemotePluginRuntime,
) error {
	if !runtime.modelsRegistrationTransferred &&
		!runtime.endpointsRegistrationTransferred &&
		!runtime.toolsRegistrationTransferred &&
		!runtime.agentStrategyRegistrationTransferred &&
		!runtime.datasourceRegistrationTransferred &&
		!runtime.triggersRegistrationTransferred {
		return errors.New("no registration transferred, cannot initialize")
	}

	files := make(map[string][]byte)
	for filename, buffer := range runtime.assets {
		files[filename] = buffer.Bytes()
	}

	// remap assets
	if err := runtime.RemapAssets(&runtime.Config, files); err != nil {
		return fmt.Errorf("assets remap failed, invalid assets data, cannot remap: %v", err)
	}

	// fill in default values
	runtime.Config.FillInDefaultValues()

	// mark assets transferred
	runtime.assetsTransferred = true

	runtime.checksum = runtime.calculateChecksum()
	runtime.InitState()
	runtime.SetActiveAt(time.Now())

	if err := runtime.Config.ManifestValidate(); err != nil {
		return fmt.Errorf("register failed, invalid manifest detected: %v", err)
	}

	// mark initialized
	runtime.initialized = true

	// spawn a core to handle CPU-intensive tasks
	routine.Submit(
		routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "debugging_runtime",
			routinepkg.RoutineLabelKeyMethod: "spawnCore",
		},
		func() { runtime.SpawnCore() },
	)

	// start heartbeat monitor
	routine.Submit(
		routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "debugging_runtime",
			routinepkg.RoutineLabelKeyMethod: "heartbeatMonitor",
		},
		func() { runtime.HeartbeatMonitor() },
	)

	return nil
}

func (d *DifyServer) handleDeclarationRegister(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	if runtime.registrationTransferred {
		return errors.New("declaration already registered")
	}

	// process handle shake if not completed
	declaration, err := parser.UnmarshalJsonBytes[plugin_entities.PluginDeclaration](registerPayload.Data)
	if err != nil {
		// close connection if handshake failed
		return fmt.Errorf("handshake failed, invalid plugin declaration: %v", err)
	}

	runtime.Config = declaration

	// registration transferred
	runtime.registrationTransferred = true

	return nil
}

func (d *DifyServer) handleToolDeclarationRegister(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	if runtime.toolsRegistrationTransferred {
		return errors.New("tools declaration already registered")
	}

	tools, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ToolProviderDeclaration](registerPayload.Data)
	if err != nil {
		return fmt.Errorf("tools register failed, invalid tools declaration: %v", err)
	}

	runtime.toolsRegistrationTransferred = true

	if len(tools) > 0 {
		declaration := runtime.Config
		declaration.Tool = &tools[0]
		runtime.Config = declaration
	}

	return nil
}

func (d *DifyServer) handleModelDeclarationRegister(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	if runtime.modelsRegistrationTransferred {
		return errors.New("models declaration already registered")
	}

	models, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ModelProviderDeclaration](registerPayload.Data)
	if err != nil {
		return fmt.Errorf("models register failed, invalid models declaration: %v", err)
	}

	runtime.modelsRegistrationTransferred = true

	if len(models) > 0 {
		declaration := runtime.Config
		declaration.Model = &models[0]
		runtime.Config = declaration
	}

	return nil
}

func (d *DifyServer) handleEndpointDeclarationRegister(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	if runtime.endpointsRegistrationTransferred {
		return errors.New("endpoints declaration already registered")
	}

	endpoints, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.EndpointProviderDeclaration](registerPayload.Data)
	if err != nil {
		return fmt.Errorf("endpoints register failed, invalid endpoints declaration: %v", err)
	}

	runtime.endpointsRegistrationTransferred = true

	if len(endpoints) > 0 {
		declaration := runtime.Config
		declaration.Endpoint = &endpoints[0]
		runtime.Config = declaration
	}

	return nil
}

func (d *DifyServer) handleAgentStrategyDeclarationRegister(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	if runtime.agentStrategyRegistrationTransferred {
		return errors.New("agent strategy declaration already registered")
	}

	agents, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.AgentStrategyProviderDeclaration](registerPayload.Data)
	if err != nil {
		return fmt.Errorf("agent strategies register failed, invalid agent strategies declaration: %v", err)
	}

	runtime.agentStrategyRegistrationTransferred = true

	if len(agents) > 0 {
		declaration := runtime.Config
		declaration.AgentStrategy = &agents[0]
		runtime.Config = declaration
	}

	return nil
}

func (d *DifyServer) handleDatasourceDeclarationRegister(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	if runtime.datasourceRegistrationTransferred {
		return errors.New("datasource declaration already registered")
	}

	datasources, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.DatasourceProviderDeclaration](registerPayload.Data)
	if err != nil {
		return fmt.Errorf("datasources register failed, invalid datasources declaration: %v", err)
	}

	runtime.datasourceRegistrationTransferred = true

	if len(datasources) > 0 {
		declaration := runtime.Config
		declaration.Datasource = &datasources[0]
		runtime.Config = declaration
	}

	return nil
}

func (d *DifyServer) handleTriggerDeclarationRegister(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) error {
	if runtime.triggersRegistrationTransferred {
		return errors.New("triggers declaration already registered")
	}

	triggers, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.TriggerProviderDeclaration](registerPayload.Data)
	if err != nil {
		return fmt.Errorf("triggers register failed, invalid triggers declaration: %v", err)
	}

	runtime.triggersRegistrationTransferred = true

	if len(triggers) > 0 {
		declaration := runtime.Config
		declaration.Trigger = &triggers[0]
		runtime.Config = declaration
	}

	return nil
}
