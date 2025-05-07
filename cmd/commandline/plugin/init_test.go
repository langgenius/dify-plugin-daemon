package plugin

import (
	"os"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/stretchr/testify/assert"
)

func TestInitPluginWithFlags(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plugin-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	oldDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(oldDir)
	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	tests := []struct {
		testName    string
		author      string
		pluginName  string
		repo        string
		description string
		quick       bool
		permissions struct {
			endpoint      bool
			tool          bool
			model         bool
			llm           bool
			textEmbedding bool
			rerank        bool
			tts           bool
			speech2text   bool
			moderation    bool
			node          bool
			app           bool
			storage       bool
		}
		expectedFiles []string
	}{
		{
			testName:    "Quick mode with minimal permissions",
			author:      "test-author",
			pluginName:  "test-plugin",
			repo:        "https://github.com/langgenius/dify-official-plugins",
			description: "Test plugin description",
			quick:       true,
			permissions: struct {
				endpoint      bool
				tool          bool
				model         bool
				llm           bool
				textEmbedding bool
				rerank        bool
				tts           bool
				speech2text   bool
				moderation    bool
				node          bool
				app           bool
				storage       bool
			}{
				endpoint: false,
				tool:     false,
				model:    false,
			},
			expectedFiles: []string{
				"test-plugin/manifest.yaml",
				"test-plugin/_assets/icon.svg",
				"test-plugin/README.md",
				"test-plugin/.env.example",
				"test-plugin/PRIVACY.md",
				"test-plugin/.github/workflows/plugin-publish.yml",
			},
		},
		{
			testName:    "Quick mode with all permissions",
			author:      "test-author",
			pluginName:  "test-plugin-full",
			repo:        "https://github.com/langgenius/dify-official-plugins",
			description: "Test plugin with all permissions",
			quick:       true,
			permissions: struct {
				endpoint      bool
				tool          bool
				model         bool
				llm           bool
				textEmbedding bool
				rerank        bool
				tts           bool
				speech2text   bool
				moderation    bool
				node          bool
				app           bool
				storage       bool
			}{
				endpoint:      true,
				tool:          true,
				model:         true,
				llm:           true,
				textEmbedding: true,
				rerank:        true,
				tts:           true,
				speech2text:   true,
				moderation:    true,
				node:          true,
				app:           true,
				storage:       true,
			},
			expectedFiles: []string{
				"test-plugin-full/manifest.yaml",
				"test-plugin-full/_assets/icon.svg",
				"test-plugin-full/README.md",
				"test-plugin-full/.env.example",
				"test-plugin-full/PRIVACY.md",
				"test-plugin-full/.github/workflows/plugin-publish.yml",
			},
		},
		{
			testName:    "Non-quick mode with tool permissions",
			author:      "test-author",
			pluginName:  "test-plugin-tool",
			repo:        "https://github.com/langgenius/dify-official-plugins",
			description: "Test plugin with tool permissions",
			quick:       false,
			permissions: struct {
				endpoint      bool
				tool          bool
				model         bool
				llm           bool
				textEmbedding bool
				rerank        bool
				tts           bool
				speech2text   bool
				moderation    bool
				node          bool
				app           bool
				storage       bool
			}{
				tool: true,
			},
			expectedFiles: []string{
				"test-plugin-tool/manifest.yaml",
				"test-plugin-tool/_assets/icon.svg",
				"test-plugin-tool/README.md",
				"test-plugin-tool/.env.example",
				"test-plugin-tool/PRIVACY.md",
				"test-plugin-tool/.github/workflows/plugin-publish.yml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Call InitPluginWithFlags with test parameters
			InitPluginWithFlags(
				tt.author,
				tt.pluginName,
				tt.repo,
				tt.description,
				tt.permissions.endpoint,
				tt.permissions.tool,
				tt.permissions.model,
				tt.permissions.llm,
				tt.permissions.textEmbedding,
				tt.permissions.rerank,
				tt.permissions.tts,
				tt.permissions.speech2text,
				tt.permissions.moderation,
				tt.permissions.node,
				tt.permissions.app,
				tt.permissions.storage,
				1024*1024, // 1MB storage
				"tool",    // category
				"python",  // language
				"0.0.1",   // min dify version
				true,      // always use quick mode
			)

			// Verify that all expected files exist
			for _, file := range tt.expectedFiles {
				_, err := os.Stat(file)
				assert.NoError(t, err, "Expected file %s to exist", file)
			}

			// If in quick mode, verify the manifest content
			if tt.quick {
				// Create a decoder for the plugin directory
				decoder, err := decoder.NewFSPluginDecoder(tt.pluginName)
				assert.NoError(t, err)
				defer decoder.Close()

				// Get the manifest
				manifest, err := decoder.Manifest()
				assert.NoError(t, err)

				// Verify basic information
				assert.Equal(t, tt.author, manifest.Author)
				assert.Equal(t, tt.pluginName, manifest.Name)
				assert.Equal(t, tt.description, manifest.Description.EnUS)

				// Verify permissions
				if tt.permissions.endpoint {
					assert.True(t, manifest.Resource.Permission.Endpoint.Enabled)
				}
				if tt.permissions.tool {
					assert.True(t, manifest.Resource.Permission.Tool.Enabled)
				}
				if tt.permissions.model {
					assert.True(t, manifest.Resource.Permission.Model.Enabled)
					if tt.permissions.llm {
						assert.True(t, manifest.Resource.Permission.Model.LLM)
					}
					if tt.permissions.textEmbedding {
						assert.True(t, manifest.Resource.Permission.Model.TextEmbedding)
					}
					if tt.permissions.rerank {
						assert.True(t, manifest.Resource.Permission.Model.Rerank)
					}
					if tt.permissions.tts {
						assert.True(t, manifest.Resource.Permission.Model.TTS)
					}
					if tt.permissions.speech2text {
						assert.True(t, manifest.Resource.Permission.Model.Speech2text)
					}
					if tt.permissions.moderation {
						assert.True(t, manifest.Resource.Permission.Model.Moderation)
					}
				}
				if tt.permissions.node {
					assert.True(t, manifest.Resource.Permission.Node.Enabled)
				}
				if tt.permissions.app {
					assert.True(t, manifest.Resource.Permission.App.Enabled)
				}
				if tt.permissions.storage {
					assert.True(t, manifest.Resource.Permission.Storage.Enabled)
					assert.Equal(t, uint64(1024*1024), manifest.Resource.Permission.Storage.Size)
				}

				// Check assets validity
				err = decoder.CheckAssetsValid()
				assert.NoError(t, err)

				// Get unique identity
				identity, err := decoder.UniqueIdentity()
				assert.NoError(t, err)
				assert.NotEmpty(t, identity)

				// Get checksum
				checksum, err := decoder.Checksum()
				assert.NoError(t, err)
				assert.NotEmpty(t, checksum)
			}
		})
	}
}

func TestInitPluginWithFlagsValidation(t *testing.T) {
	tests := []struct {
		name        string
		author      string
		pluginName  string
		repo        string
		description string
		expectError bool
	}{
		{
			name:        "Valid inputs",
			author:      "test-author",
			pluginName:  "test-plugin",
			repo:        "https://github.com/langgenius/dify-official-plugins",
			description: "Test description",
			expectError: false,
		},
		{
			name:        "Invalid author (uppercase)",
			author:      "Test-Author",
			pluginName:  "test-plugin",
			repo:        "https://github.com/langgenius/dify-official-plugins",
			description: "Test description",
			expectError: true,
		},
		{
			name:        "Invalid plugin name (uppercase)",
			author:      "test-author",
			pluginName:  "Test-Plugin",
			repo:        "https://github.com/langgenius/dify-official-plugins",
			description: "Test description",
			expectError: true,
		},
		{
			name:        "Empty description",
			author:      "test-author",
			pluginName:  "test-plugin",
			repo:        "https://github.com/langgenius/dify-official-plugins",
			description: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tempDir, err := os.MkdirTemp("", "plugin-test-*")
			assert.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Change to the temporary directory
			oldDir, err := os.Getwd()
			assert.NoError(t, err)
			defer os.Chdir(oldDir)
			err = os.Chdir(tempDir)
			assert.NoError(t, err)

			// Call InitPluginWithFlags
			InitPluginWithFlags(
				tt.author,
				tt.pluginName,
				tt.repo,
				tt.description,
				false, // allowRegisterEndpoint
				false, // allowInvokeTool
				false, // allowInvokeModel
				false, // allowInvokeLLM
				false, // allowInvokeTextEmbedding
				false, // allowInvokeRerank
				false, // allowInvokeTTS
				false, // allowInvokeSpeech2Text
				false, // allowInvokeModeration
				false, // allowInvokeNode
				false, // allowInvokeApp
				false, // allowUseStorage
				0,     // storageSize
				"",    // category
				"",    // language
				"",    // minDifyVersion
				true,  // quick
			)

			// Check if the plugin directory was created
			_, err = os.Stat(tt.pluginName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
