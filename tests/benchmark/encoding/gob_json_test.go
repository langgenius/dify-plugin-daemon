package encoding

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/tests"
	"github.com/vmihailenco/msgpack/v5"
)

func BenchmarkMsgpackVsJson(b *testing.B) {
	declaration := plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Version: "0.0.1",
			Type:    manifest_entities.PluginType,
			Description: plugin_entities.I18nObject{
				EnUS: "test",
			},
			Name: "test",
			Icon: "test.svg",
			Label: plugin_entities.I18nObject{
				EnUS: "test",
			},
			Author:    "test",
			CreatedAt: time.Now(),
			Resource: plugin_entities.PluginResourceRequirement{
				Memory: 1,
				Permission: &plugin_entities.PluginPermissionRequirement{
					Tool: &plugin_entities.PluginPermissionToolRequirement{
						Enabled: true,
					},
					Model: &plugin_entities.PluginPermissionModelRequirement{
						Enabled: true,
					},
					Node: &plugin_entities.PluginPermissionNodeRequirement{
						Enabled: true,
					},
					Storage: &plugin_entities.PluginPermissionStorageRequirement{
						Enabled: true,
						Size:    1024,
					},
				},
			},
			Plugins: plugin_entities.PluginExtensions{},
			Meta: plugin_entities.PluginMeta{
				Version: "0.0.1",
				Arch: []constants.Arch{
					constants.AMD64,
				},
				Runner: plugin_entities.PluginRunner{
					Language:   constants.Python,
					Version:    "3.12",
					Entrypoint: "main",
				},
			},
		},
		Model: &plugin_entities.ModelProviderDeclaration{
			Provider: "test",
			Label: plugin_entities.I18nObject{
				EnUS: "test",
			},
			Description: &plugin_entities.I18nObject{
				EnUS: "test",
			},
			IconSmall: &plugin_entities.I18nObject{
				EnUS: "test",
			},
			IconLarge: &plugin_entities.I18nObject{
				EnUS: "test",
			},
			ProviderCredentialSchema: &plugin_entities.ModelProviderCredentialSchema{
				CredentialFormSchemas: []plugin_entities.ModelProviderCredentialFormSchema{
					{
						Variable: "test",
						Label: plugin_entities.I18nObject{
							EnUS: "test",
						},
					},
				},
			},
			Models: []plugin_entities.ModelDeclaration{},
		},
	}

	// add 100 models to the declaration
	for i := 0; i < 100; i++ {
		declaration.Model.Models = append(declaration.Model.Models, plugin_entities.ModelDeclaration{
			Model: "test",
			Label: plugin_entities.I18nObject{
				EnUS: "test",
			},
			ModelProperties: map[string]any{
				"test": "test",
			},
			ParameterRules: []plugin_entities.ModelParameterRule{
				{
					Name: "test",
					Label: &plugin_entities.I18nObject{
						EnUS: "test",
					},
					Type:     parser.ToPtr(plugin_entities.PARAMETER_TYPE_BOOLEAN),
					Required: true,
					Help: &plugin_entities.I18nObject{
						EnUS: "test",
					},
				},
				{
					Name: "test1",
					Label: &plugin_entities.I18nObject{
						EnUS: "test",
					},
					Type:     parser.ToPtr(plugin_entities.PARAMETER_TYPE_BOOLEAN),
					Required: true,
					Help: &plugin_entities.I18nObject{
						EnUS: "test",
					},
				},
				{
					Name: "test2",
					Label: &plugin_entities.I18nObject{
						EnUS: "test",
					},
					Type:     parser.ToPtr(plugin_entities.PARAMETER_TYPE_BOOLEAN),
					Required: true,
					Help: &plugin_entities.I18nObject{
						EnUS: "test",
					},
				},
				{
					Name: "test3",
					Label: &plugin_entities.I18nObject{
						EnUS: "test",
					},
					Type:     parser.ToPtr(plugin_entities.PARAMETER_TYPE_BOOLEAN),
					Required: true,
					Help: &plugin_entities.I18nObject{
						EnUS: "test",
					},
				},
			},
		})
	}

	var msgpackBytes []byte
	var jsonBytes []byte
	var cborBytes []byte
	var gobBytes []byte
	totalBytes := 0

	// Encode benchmarks
	b.Run("Msgpack Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var err error
			msgpackBytes, err = msgpack.Marshal(declaration)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(msgpackBytes)
		}
	})
	b.Log("Msgpack encoded size:", tests.ReadableBytes(len(msgpackBytes)))
	b.Log("Total bytes encoded with Msgpack:", tests.ReadableBytes(totalBytes))

	totalBytes = 0
	b.Run("Json Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var err error
			jsonBytes, err = json.Marshal(declaration)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(jsonBytes)
		}
	})
	b.Log("Json encoded size:", tests.ReadableBytes(len(jsonBytes)))
	b.Log("Total bytes encoded with Json:", tests.ReadableBytes(totalBytes))

	totalBytes = 0
	b.Run("CBOR Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var err error
			cborBytes, err = cbor.Marshal(declaration)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(cborBytes)
		}
	})
	b.Log("CBOR encoded size:", tests.ReadableBytes(len(cborBytes)))
	b.Log("Total bytes encoded with CBOR:", tests.ReadableBytes(totalBytes))

	totalBytes = 0
	b.Run("GOB Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var err error
			var buffer bytes.Buffer
			enc := gob.NewEncoder(&buffer)
			err = enc.Encode(declaration)
			if err != nil {
				b.Fatal(err)
			}
			gobBytes = buffer.Bytes()
			totalBytes += len(gobBytes)
		}
	})
	b.Log("GOB encoded size:", tests.ReadableBytes(len(gobBytes)))
	b.Log("Total bytes encoded with GOB:", tests.ReadableBytes(totalBytes))

	// Decode benchmarks
	totalBytes = 0
	b.Run("Msgpack Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := msgpack.Unmarshal(msgpackBytes, &declaration)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(msgpackBytes)
		}
	})
	b.Log("Total bytes decoded with Msgpack:", tests.ReadableBytes(totalBytes))

	totalBytes = 0
	b.Run("Json Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var err error
			var decodedDeclaration plugin_entities.PluginDeclaration
			err = json.Unmarshal(jsonBytes, &decodedDeclaration)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(jsonBytes)
		}
	})
	b.Log("Total bytes decoded with Json:", tests.ReadableBytes(totalBytes))

	totalBytes = 0
	b.Run("CBOR Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var decodedDeclaration plugin_entities.PluginDeclaration
			err := cbor.Unmarshal(cborBytes, &decodedDeclaration)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(cborBytes)
		}
	})
	b.Log("Total bytes decoded with CBOR:", tests.ReadableBytes(totalBytes))

	totalBytes = 0
	b.Run("GOB Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var decodedDeclaration plugin_entities.PluginDeclaration
			dec := gob.NewDecoder(bytes.NewBuffer(gobBytes))
			err := dec.Decode(&decodedDeclaration)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(gobBytes)
		}
	})
	b.Log("Total bytes decoded with GOB:", tests.ReadableBytes(totalBytes))

	totalBytes = 0
	b.Run("Map Json Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var decoded map[string]any
			err := json.Unmarshal(jsonBytes, &decoded)
			if err != nil {
				b.Fatal(err)
			}
			totalBytes += len(jsonBytes)
		}
	})
	b.Log("Total map bytes decoded with Json:", tests.ReadableBytes(totalBytes))
}
