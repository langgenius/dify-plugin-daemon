package plugin_entities

import (
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
)

func preparePluginDeclaration() PluginDeclaration {
	return PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: PluginDeclarationWithoutAdvancedFields{
			Version: "0.0.1",
			Type:    manifest_entities.PluginType,
			Description: I18nObject{
				EnUS: "test",
			},
			Name: "test",
			Icon: "test.svg",
			Label: I18nObject{
				EnUS: "test",
			},
			Author:    "test",
			CreatedAt: time.Now(),
			Resource: PluginResourceRequirement{
				Memory: 1,
				Permission: &PluginPermissionRequirement{
					Tool: &PluginPermissionToolRequirement{
						Enabled: true,
					},
					Model: &PluginPermissionModelRequirement{
						Enabled: true,
					},
					Node: &PluginPermissionNodeRequirement{
						Enabled: true,
					},
					Storage: &PluginPermissionStorageRequirement{
						Enabled: true,
						Size:    1024,
					},
				},
			},
			Plugins: PluginExtensions{},
			Meta: PluginMeta{
				Version: "0.0.1",
				Arch: []constants.Arch{
					constants.AMD64,
				},
				Runner: PluginRunner{
					Language:   constants.Python,
					Version:    "3.12",
					Entrypoint: "main",
				},
			},
		},
	}
}

func TestPluginDeclarationFullTest(t *testing.T) {
	declaration := preparePluginDeclaration()
	declarationBytes := parser.MarshalJsonBytes(declaration)

	// unmarshal
	newDeclaration, err := parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err != nil {
		t.Errorf("failed to unmarshal declaration: %s", err.Error())
		return
	}

	if newDeclaration.Version != declaration.Version {
		t.Errorf("version not equal")
		return
	}
	if newDeclaration.Type != declaration.Type {
		t.Errorf("type not equal")
		return
	}
	if newDeclaration.Name != declaration.Name {
		t.Errorf("name not equal")
		return
	}
	if newDeclaration.Author != declaration.Author {
		t.Errorf("author not equal")
		return
	}
	if newDeclaration.Resource.Memory != declaration.Resource.Memory {
		t.Errorf("memory not equal")
		return
	}

	if newDeclaration.Resource.Permission == nil {
		t.Errorf("permission is nil")
		return
	}

	if newDeclaration.Resource.Permission.Tool == nil {
		t.Errorf("tool permission is nil")
		return
	}

	if newDeclaration.Resource.Permission.Node == nil {
		t.Errorf("node permission is nil")
		return
	}

	if newDeclaration.Resource.Permission.Storage == nil {
		t.Errorf("storage permission is nil")
		return
	}
}

func TestPluginDeclarationIncorrectVersion(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Version = "1"
	declarationBytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err == nil {
		t.Errorf("failed to validate version")
		return
	}
}

func TestPluginUnsupportedLanguage(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Meta.Runner.Language = "test"
	declarationBytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err == nil {
		t.Errorf("failed to validate language")
		return
	}
}

func TestPluginUnsupportedArch(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Meta.Arch[0] = constants.Arch("test")
	declarationBytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err == nil {
		t.Errorf("failed to validate arch")
		return
	}
}

func TestPluginStorageSizeTooSmallOrTooLarge(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Resource.Permission.Storage.Size = 1023
	declarationBytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err == nil {
		t.Errorf("failed to validate storage size")
		return
	}

	declaration.Resource.Permission.Storage.Size = 1073741825
	declarationBytes = parser.MarshalJsonBytes(declaration)

	_, err = parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err == nil {
		t.Errorf("failed to validate storage size")
		return
	}
}

func TestPluginDeclarationIncorrectType(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Type = "test"
	declarationBytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err == nil {
		t.Errorf("failed to validate type")
		return
	}
}

func TestPluginDeclarationIncorrectName(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Name = ""
	declarationBytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declarationBytes)
	if err == nil {
		t.Errorf("failed to validate name")
		return
	}
}
