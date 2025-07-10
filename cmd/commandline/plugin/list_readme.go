package plugin

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// Language represents supported README languages
type Language struct {
	Code      string
	Name      string
	Available bool
}

// GetLanguageName returns the full language name for a given language code
func GetLanguageName(code string) string {
	languageNames := map[string]string{
		"en_US":   "English",
		"zh_Hans": "简体中文 (Simplified Chinese)",
		"ja_JP":   "日本語 (Japanese)",
		"pt_BR":   "Português (Portuguese - Brazil)",
	}

	if name, exists := languageNames[code]; exists {
		return name
	}
	return "unknown"
}

// ListReadme displays README language information in table format for a specific plugin
func ListReadme(pluginPath string) {
	var pluginDecoder decoder.PluginDecoder
	var err error

	stat, err := os.Stat(pluginPath)
	if err != nil {
		log.Error("failed to get plugin file stat: %s", err)
		return
	}

	if stat.IsDir() {
		pluginDecoder, err = decoder.NewFSPluginDecoder(pluginPath)
	} else {
		fileContent, err := os.ReadFile(pluginPath)
		if err != nil {
			log.Error("failed to read plugin file: %s", err)
			return
		}
		pluginDecoder, err = decoder.NewZipPluginDecoder(fileContent)
		if err != nil {
			log.Error("failed to create zip plugin decoder: %s", err)
			return
		}
	}
	if err != nil {
		log.Error("your plugin is not a valid plugin: %s", err)
		return
	}

	// Get available i18n README files
	availableReadmes, err := pluginDecoder.AvailableI18nReadme()
	if err != nil {
		log.Error("failed to get available README files: %s", err)
		return
	}

	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 3, ' ', 0)

	// Print table header
	fmt.Fprintln(w, "language-code\tlanguage\tavailable")
	fmt.Fprintln(w, "-------------\t--------\t---------")

	// Print each available README
	for code, _ := range availableReadmes {
		languageName := GetLanguageName(code)
		fmt.Fprintf(w, "%s\t%s\t✅\n", code, languageName)
	}

	// Flush the writer to ensure all output is printed
	w.Flush()
}
