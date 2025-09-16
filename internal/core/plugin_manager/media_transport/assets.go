package media_transport

import (
	"errors"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (m *MediaBucket) RemapAssets(declaration *plugin_entities.PluginDeclaration, assets map[string][]byte) ([]string, error) {
	remappedAssetIds := make(map[string]string)
	assetsIds := []string{}
	remap := func(filename string) (string, error) {
		if id, ok := remappedAssetIds[filename]; ok {
			return id, nil
		}

		file, ok := assets[filename]
		if !ok {
			return "", fmt.Errorf("file not found: %s", filename)
		}

		id, err := m.Upload(filename, file)
		if err != nil {
			return "", err
		}

		assetsIds = append(assetsIds, id)

		remappedAssetIds[filename] = id
		return id, nil
	}

	var err error

	if declaration.Model != nil {
		iconFields := []struct {
			icon      *plugin_entities.I18nObject
			iconType  string
			fieldName string
		}{
			{declaration.Model.IconSmall, "model icon small", ""},
			{declaration.Model.IconLarge, "model icon large", ""},
			{declaration.Model.IconSmallDark, "model icon small dark", ""},
			{declaration.Model.IconLargeDark, "model icon large dark", ""},
		}

		langFields := []struct {
			get    func(*plugin_entities.I18nObject) *string
			suffix string
		}{
			{func(i *plugin_entities.I18nObject) *string { return &i.EnUS }, "en_US"},
			{func(i *plugin_entities.I18nObject) *string { return &i.ZhHans }, "zh_Hans"},
			{func(i *plugin_entities.I18nObject) *string { return &i.JaJp }, "ja_JP"},
			{func(i *plugin_entities.I18nObject) *string { return &i.PtBr }, "pt_BR"},
		}

		for _, iconField := range iconFields {
			if iconField.icon == nil {
				continue
			}
			for _, langField := range langFields {
				valPtr := langField.get(iconField.icon)
				if valPtr != nil && *valPtr != "" {
					*valPtr, err = remap(*valPtr)
					if err != nil {
						return nil, errors.Join(err, fmt.Errorf("failed to remap %s %s", iconField.iconType, langField.suffix))
					}
				}
			}
		}
	}

	if declaration.Tool != nil {
		if declaration.Tool.Identity.Icon != "" {
			declaration.Tool.Identity.Icon, err = remap(declaration.Tool.Identity.Icon)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to remap tool icon"))
			}
		}

		if declaration.Tool.Identity.IconDark != "" {
			declaration.Tool.Identity.IconDark, err = remap(declaration.Tool.Identity.IconDark)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to remap tool icon dark"))
			}
		}
	}

	if declaration.AgentStrategy != nil {
		if declaration.AgentStrategy.Identity.Icon != "" {
			declaration.AgentStrategy.Identity.Icon, err = remap(declaration.AgentStrategy.Identity.Icon)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to remap agent icon"))
			}
		}

		if declaration.AgentStrategy.Identity.IconDark != "" {
			declaration.AgentStrategy.Identity.IconDark, err = remap(declaration.AgentStrategy.Identity.IconDark)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to remap agent icon dark"))
			}
		}
	}

	if declaration.Datasource != nil {
		if declaration.Datasource.Identity.Icon != "" {
			declaration.Datasource.Identity.Icon, err = remap(declaration.Datasource.Identity.Icon)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to remap datasource icon"))
			}
		}

		for i, datasource := range declaration.Datasource.Datasources {
			if datasource.Identity.Icon != "" {
				remappedIcon, err := remap(datasource.Identity.Icon)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap datasource step icon"))
				}

				identityPointer := &declaration.Datasource.Datasources[i].Identity
				identityPointer.Icon = remappedIcon
			}
		}
	}

	if declaration.Icon != "" {
		declaration.Icon, err = remap(declaration.Icon)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("failed to remap plugin icon"))
		}
	}

	if declaration.IconDark != "" {
		declaration.IconDark, err = remap(declaration.IconDark)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("failed to remap plugin dark icon"))
		}
	}

	return assetsIds, nil
}
