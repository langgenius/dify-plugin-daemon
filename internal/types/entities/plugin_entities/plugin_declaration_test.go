package plugin_entities

import (
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func preparePluginDeclaration() PluginDeclaration {
	return PluginDeclaration{
		Version:   "0.0.1",
		Type:      PluginType,
		Name:      "test",
		Author:    "test",
		CreatedAt: time.Now(),
		Resource: PluginResourceRequirement{
			Memory:  1,
			Storage: 1,
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
			},
		},
		Plugins: []string{},
		Execution: PluginExecution{
			Install: "echo 'hello'",
			Launch:  "echo 'hello'",
		},
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
	}
}

func TestPluginDeclarationFullTest(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration_bytes := parser.MarshalJsonBytes(declaration)

	// unmarshal
	new_declaration, err := parser.UnmarshalJsonBytes[PluginDeclaration](declaration_bytes)
	if err != nil {
		t.Errorf("failed to unmarshal declaration: %s", err.Error())
		return
	}

	if new_declaration.Version != declaration.Version {
		t.Errorf("version not equal")
		return
	}
	if new_declaration.Type != declaration.Type {
		t.Errorf("type not equal")
		return
	}
	if new_declaration.Name != declaration.Name {
		t.Errorf("name not equal")
		return
	}
	if new_declaration.Author != declaration.Author {
		t.Errorf("author not equal")
		return
	}
	if new_declaration.Resource.Memory != declaration.Resource.Memory {
		t.Errorf("memory not equal")
		return
	}

	if new_declaration.Resource.Permission == nil {
		t.Errorf("permission is nil")
		return
	}

	if new_declaration.Resource.Permission.Tool == nil {
		t.Errorf("tool permission is nil")
		return
	}

	if new_declaration.Resource.Permission.Node == nil {
		t.Errorf("node permission is nil")
		return
	}

}

func TestPluginDeclarationIncorrectVersion(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Version = "1"
	declaration_bytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declaration_bytes)
	if err == nil {
		t.Errorf("failed to validate version")
		return
	}
}

func TestPluginUnsupportedLanguage(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Meta.Runner.Language = "test"
	declaration_bytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declaration_bytes)
	if err == nil {
		t.Errorf("failed to validate language")
		return
	}
}

func TestPluginUnsupportedArch(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Meta.Arch[0] = constants.Arch("test")
	declaration_bytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declaration_bytes)
	if err == nil {
		t.Errorf("failed to validate arch")
		return
	}
}

func TestPluginDeclarationIncorrectType(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Type = "test"
	declaration_bytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declaration_bytes)
	if err == nil {
		t.Errorf("failed to validate type")
		return
	}
}

func TestPluginDeclarationIncorrectName(t *testing.T) {
	declaration := preparePluginDeclaration()
	declaration.Name = ""
	declaration_bytes := parser.MarshalJsonBytes(declaration)

	_, err := parser.UnmarshalJsonBytes[PluginDeclaration](declaration_bytes)
	if err == nil {
		t.Errorf("failed to validate name")
		return
	}
}
