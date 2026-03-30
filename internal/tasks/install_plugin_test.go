package tasks

import (
	"errors"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestResolveNewDeclaration(t *testing.T) {
	existingDecl := &plugin_entities.PluginDeclaration{}
	fetchedDecl := &plugin_entities.PluginDeclaration{}
	fetchErr := errors.New("declaration not found")

	noopFetch := func(_ plugin_entities.PluginUniqueIdentifier, _ plugin_entities.PluginRuntimeType) (*plugin_entities.PluginDeclaration, error) {
		t.Error("fetch should not be called when declaration is already set")
		return nil, nil
	}

	tests := []struct {
		name    string
		decl    *plugin_entities.PluginDeclaration
		fetch   func(plugin_entities.PluginUniqueIdentifier, plugin_entities.PluginRuntimeType) (*plugin_entities.PluginDeclaration, error)
		want    *plugin_entities.PluginDeclaration
		wantErr bool
	}{
		{
			name:  "declaration already set — fetcher not called",
			decl:  existingDecl,
			fetch: noopFetch,
			want:  existingDecl,
		},
		{
			name: "nil declaration — fetcher called and succeeds",
			decl: nil,
			fetch: func(_ plugin_entities.PluginUniqueIdentifier, _ plugin_entities.PluginRuntimeType) (*plugin_entities.PluginDeclaration, error) {
				return fetchedDecl, nil
			},
			want: fetchedDecl,
		},
		{
			name: "nil declaration — fetcher returns error",
			decl: nil,
			fetch: func(_ plugin_entities.PluginUniqueIdentifier, _ plugin_entities.PluginRuntimeType) (*plugin_entities.PluginDeclaration, error) {
				return nil, fetchErr
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveNewDeclaration(tt.decl, "author/plugin:1.0.0", plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL, tt.fetch)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveNewDeclaration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveNewDeclaration() = %v, want %v", got, tt.want)
			}
		})
	}
}
