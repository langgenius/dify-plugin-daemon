package service

import (
	"errors"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"gorm.io/gorm"
)

func GetPluginReadmeMap(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (map[string]string, error) {
	readmeMap, err := getPluginReadmeMapFromDb(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}
	if readmeMap == nil {
		readmeMap = make(map[string]string)
	}
	if len(readmeMap) == 0 {
		readmeMap, err = extractInstalledPluginReadmeMap(pluginUniqueIdentifier)
		if err != nil {
			return nil, err
		}
		err := savePluginReadmeMapToDb(pluginUniqueIdentifier, readmeMap)
		if err != nil {
			return nil, err
		}
	}
	if len(readmeMap) == 0 {
		return nil, nil
	}
	return readmeMap, nil
}

func getPluginReadmeMapFromDb(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (map[string]string, error) {
	readmes, err := db.GetAll[models.PluginReadmeRecord](
		db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
	)
	if err != nil {
		return nil, err
	}
	readmeMap := make(map[string]string)
	for _, readme := range readmes {
		readmeMap[readme.Language] = readme.Content
	}
	return readmeMap, nil
}

func extractInstalledPluginReadmeMap(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (map[string]string, error) {
	manager := plugin_manager.Manager()
	pkgBytes, err := manager.GetPackage(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}

	zipDecoder, err := decoder.NewZipPluginDecoder(pkgBytes)
	if err != nil {
		return nil, err
	}

	readmeMap, err := zipDecoder.AvailableI18nReadme()
	if err != nil {
		return nil, err
	}

	if readmeMap == nil {
		readmeMap = make(map[string]string)
	}
	return readmeMap, nil
}

func savePluginReadmeMapToDb(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	readmeMap map[string]string,
) error {
	return db.WithTransaction(func(tx *gorm.DB) error {
		// Create new readme entries
		for language, content := range readmeMap {
			readme := models.PluginReadmeRecord{
				PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
				Language:               language,
				Content:                content,
			}
			return db.Create(&readme, tx)
		}

		return nil
	})
}

func FetchPluginReadme(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	language string,
) *entities.Response {
	readmeMap, err := GetPluginReadmeMap(pluginUniqueIdentifier)
	if err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to get readme from database: %w", err)).ToResponse()
	}

	if len(readmeMap) == 0 {
		return exception.NotFoundError(errors.New("no readme content available for this plugin")).ToResponse()
	}

	var selectedContent string
	var selectedLanguage string

	if content, exists := readmeMap[language]; exists {
		selectedContent = content
		selectedLanguage = language
	} else if content, exists := readmeMap["en_US"]; exists {
		selectedContent = content
		selectedLanguage = "en_US"
	} else {
		for lang, content := range readmeMap {
			selectedContent = content
			selectedLanguage = lang
			break
		}
	}

	return entities.NewSuccessResponse(map[string]interface{}{
		"content":  selectedContent,
		"language": selectedLanguage,
	})
}
