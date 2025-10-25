package debugging_runtime

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (d *DifyServer) handleHandleShake(
	runtime *RemotePluginRuntime,
	registerPayload plugin_entities.RemotePluginRegisterPayload,
) (*ConnectionInfo, error) {
	if runtime.handshake {
		return nil, errors.New("handshake already completed")
	}

	key, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterHandshake](registerPayload.Data)
	if err != nil {
		// close connection if handshake failed
		return nil, errors.New("handshake failed, invalid handshake message")
	}

	info, err := GetConnectionInfo(key.Key)
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
		!runtime.datasourceRegistrationTransferred {
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

	// trigger registration event
	if err := runtime.Register(); err != nil {
		return fmt.Errorf(fmt.Sprintf("register failed, cannot register: %v", err))
	}

	if err := runtime.Config.ManifestValidate(); err != nil {
		return fmt.Errorf("register failed, invalid manifest detected: %v", err)
	}

	// mark initialized
	runtime.initialized = true

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
