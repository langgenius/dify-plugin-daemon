package plugin_entities

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Synthetic declarations isolate encoding cost; they are not production captures.
func declarationJSONFixture(models int) *PluginDeclaration {
	p := preparePluginDeclaration()
	p.CreatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p.Description.EnUS = "Unicode 中文 <tag> & \"quoted\"\n"
	p.Model = &ModelProviderDeclaration{Provider: "example"}
	for i := 0; i < models; i++ {
		label := I18nObject{}
		if i%2 != 0 {
			label.EnUS = "Explicit label"
		}
		p.Model.Models = append(p.Model.Models, ModelDeclaration{
			Model: fmt.Sprintf("model-%d", i), Label: label, ModelType: "llm",
			ModelProperties: map[string]any{
				"context_size": 128000, "description": strings.Repeat("model description 中文 <>& ", 16),
			},
		})
	}
	p.FillInDefaultValues()
	return &p
}

func TestDeclarationDefaultsBeforeEncoding(t *testing.T) {
	for _, decode := range []struct {
		name string
		fn   func([]byte, any) error
	}{
		{"json", json.Unmarshal},
		{"yaml", yaml.Unmarshal}, // JSON is also valid YAML.
	} {
		t.Run(decode.name, func(t *testing.T) {
			var model ModelDeclaration
			require.NoError(t, decode.fn([]byte(`{"model":"fallback","label":{"zh_Hans":"中文"}}`), &model))
			require.Equal(t, "fallback", model.Label.EnUS)
			require.Equal(t, "中文", model.Label.ZhHans)
			for _, models := range []string{`null`, `[]`, `{"llm":{"predefined":[]}}`} {
				var provider ModelProviderDeclaration
				require.NoError(t, decode.fn([]byte(`{"provider":"example","models":`+models+`}`), &provider))
				require.NotNil(t, provider.Models)
			}
		})
	}
	var plugin PluginDeclaration
	require.NoError(t, json.Unmarshal([]byte(`{"description":{"en_US":"description"},"model":{"provider":"example","models":[{"model":"fallback"},{"model":"other","label":{"en_US":"explicit"}}]}}`), &plugin))
	require.NotNil(t, plugin.Tags)
	require.NotNil(t, plugin.Model.Description)
	require.Equal(t, "description", plugin.Model.Description.EnUS)
	require.Equal(t, "fallback", plugin.Model.Models[0].Label.EnUS)
	require.Equal(t, "explicit", plugin.Model.Models[1].Label.EnUS)
	before, err := json.Marshal(&plugin)
	require.NoError(t, err)
	plugin.FillInDefaultValues()
	after, err := json.Marshal(&plugin)
	require.NoError(t, err)
	require.Equal(t, string(before), string(after), "default initialization must be idempotent")
}

func TestDeclarationConcurrentJSONIsReadOnly(t *testing.T) {
	p := declarationJSONFixture(10)
	before, err := parser.MarshalCBOR(p)
	require.NoError(t, err)
	want, err := json.Marshal(p)
	require.NoError(t, err)
	var workers sync.WaitGroup
	for i := 0; i < 16; i++ {
		workers.Go(func() {
			for j := 0; j < 20; j++ {
				got, err := json.Marshal(p)
				if err != nil || string(got) != string(want) {
					t.Errorf("concurrent encoding changed output: %v", err)
					return
				}
			}
		})
	}
	workers.Wait()
	after, err := parser.MarshalCBOR(p)
	require.NoError(t, err)
	// Compare decoded values because CBOR map iteration order is unspecified.
	beforeValue, err := parser.UnmarshalCBOR[PluginDeclaration](before)
	require.NoError(t, err)
	afterValue, err := parser.UnmarshalCBOR[PluginDeclaration](after)
	require.NoError(t, err)
	require.Equal(t, beforeValue, afterValue)
}

func TestDeclarationUsesStandardJSONEncoding(t *testing.T) {
	for _, value := range []any{PluginDeclaration{}, &PluginDeclaration{}, ModelProviderDeclaration{}, &ModelProviderDeclaration{}, ModelDeclaration{}, &ModelDeclaration{}} {
		_, custom := value.(json.Marshaler)
		require.False(t, custom, "%T must use the outer JSON encoder", value)
	}
}

func declarationJSONCases() map[string]any {
	model := declarationJSONFixture(2)
	tool := declarationJSONFixture(0)
	tool.Model = nil
	tool.Tool = &ToolProviderDeclaration{}
	tool.FillInDefaultValues()
	trigger := declarationJSONFixture(0)
	trigger.Model = nil
	trigger.Trigger = &TriggerProviderDeclaration{}
	trigger.FillInDefaultValues()
	return map[string]any{
		"manifest":     entities.NewSuccessResponse(model),
		"empty-models": entities.NewSuccessResponse(declarationJSONFixture(0)),
		"list": entities.NewSuccessResponse(struct {
			List  []*PluginDeclaration `json:"list"`
			Total int                  `json:"total"`
		}{[]*PluginDeclaration{model, tool, trigger}, 3}),
		"provider": model.Model,
		"model":    model.Model.Models[0],
	}
}

// Goldens were captured before removing the nested MarshalJSON methods.
func TestDeclarationJSONCompatibility(t *testing.T) {
	for name, response := range declarationJSONCases() {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.JSON(http.StatusOK, response)
			require.Equal(t, http.StatusOK, w.Code)
			path := filepath.Join("testdata", "declaration_json", name+".json")
			want, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, string(want), w.Body.String())
		})
	}
}

func BenchmarkDeclarationJSONParallel(b *testing.B) {
	gin.SetMode(gin.TestMode)
	p := declarationJSONFixture(100)
	type installation struct {
		Declaration *PluginDeclaration `json:"declaration"`
	}
	list := make([]installation, 32)
	for i := range list {
		list[i].Declaration = p
	}
	response := entities.NewSuccessResponse(struct {
		List  []installation `json:"list"`
		Total int            `json:"total"`
	}{list, len(list)})
	encoded, err := json.Marshal(response)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(encoded)))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		w := &declarationJSONWriter{header: make(http.Header)}
		c, _ := gin.CreateTestContext(w)
		for pb.Next() {
			w.bytes = 0
			c.JSON(http.StatusOK, response)
			if w.bytes != len(encoded) {
				b.Error("incomplete JSON response")
				return
			}
		}
	})
}

type declarationJSONWriter struct {
	header http.Header
	bytes  int
}

func (w *declarationJSONWriter) Header() http.Header { return w.header }
func (w *declarationJSONWriter) WriteHeader(int)     {}
func (w *declarationJSONWriter) Write(p []byte) (int, error) {
	w.bytes += len(p)
	return len(p), nil
}

// Exercise c.JSON without retaining a second copy in a ResponseRecorder.
// Construction and normalization happen before timing, as on a cache hit.
func BenchmarkDeclarationJSON(b *testing.B) {
	gin.SetMode(gin.TestMode)
	for _, models := range []int{10, 100} {
		for _, page := range []int{0, 1, 32, 256} {
			name := fmt.Sprintf("models=%d/page=%d", models, page)
			b.Run(name, func(b *testing.B) {
				p := declarationJSONFixture(models)
				response := entities.NewSuccessResponse(p)
				if page > 0 {
					type installation struct {
						ID          string             `json:"id"`
						Declaration *PluginDeclaration `json:"declaration"`
					}
					list := make([]installation, page)
					for i := range list {
						list[i] = installation{fmt.Sprint(i), p}
					}
					response = entities.NewSuccessResponse(struct {
						List  []installation `json:"list"`
						Total int            `json:"total"`
					}{list, page})
				}
				encoded, err := json.Marshal(response)
				if err != nil {
					b.Fatal(err)
				}
				b.SetBytes(int64(len(encoded)))
				b.ReportAllocs()
				w := &declarationJSONWriter{header: make(http.Header)}
				c, _ := gin.CreateTestContext(w)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					w.bytes = 0
					c.JSON(http.StatusOK, response)
					if w.bytes != len(encoded) {
						b.Fatal("incomplete JSON response")
					}
				}
			})
		}
	}
}
