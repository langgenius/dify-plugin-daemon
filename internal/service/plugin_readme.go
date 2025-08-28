package service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func GetPluginReadmeMap(
	tenantId string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (map[string]string, error) {
	readmeMap, err := getPluginReadmeMapFromDb(tenantId, pluginUniqueIdentifier)
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
		err := savePluginReadmeMapToDb(tenantId, pluginUniqueIdentifier, readmeMap)
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
	tenantId string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (map[string]string, error) {
	var readmes []models.PluginReadme
	err := db.DifyPluginDB.Where(
		"tenant_id = ? AND plugin_unique_identifier = ?",
		tenantId, pluginUniqueIdentifier.String(),
	).Find(&readmes).Error
	if err != nil {
		return nil, err
	}
	if readmes == nil {
		return nil, nil
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
	tenantId string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	readmeMap map[string]string,
) error {
	tx := db.DifyPluginDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create new readme entries
	for language, content := range readmeMap {
		readme := models.PluginReadme{
			TenantID:               tenantId,
			PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
			Language:               language,
			Content:                content,
		}
		if err := tx.Create(&readme).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
