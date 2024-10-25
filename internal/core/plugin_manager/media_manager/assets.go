package media_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (m *MediaManager) RemapAssets(declaration *plugin_entities.PluginDeclaration, assets map[string][]byte) ([]string, error) {
	remapped_asset_ids := make(map[string]string)
	assets_ids := []string{}
	remap := func(filename string) (string, error) {
		if id, ok := remapped_asset_ids[filename]; ok {
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

		assets_ids = append(assets_ids, id)

		remapped_asset_ids[filename] = id
		return id, nil
	}

	var err error

	if declaration.Model != nil {
		if declaration.Model.IconSmall != nil {
			if declaration.Model.IconSmall.EnUS != "" {
				declaration.Model.IconSmall.EnUS, err = remap(declaration.Model.IconSmall.EnUS)
				if err != nil {
					return nil, err
				}
			}

			if declaration.Model.IconSmall.ZhHans != "" {
				declaration.Model.IconSmall.ZhHans, err = remap(declaration.Model.IconSmall.ZhHans)
				if err != nil {
					return nil, err
				}
			}

			if declaration.Model.IconSmall.JaJp != "" {
				declaration.Model.IconSmall.JaJp, err = remap(declaration.Model.IconSmall.JaJp)
				if err != nil {
					return nil, err
				}
			}

			if declaration.Model.IconSmall.PtBr != "" {
				declaration.Model.IconSmall.PtBr, err = remap(declaration.Model.IconSmall.PtBr)
				if err != nil {
					return nil, err
				}
			}
		}

		if declaration.Model.IconLarge != nil {
			if declaration.Model.IconLarge.EnUS != "" {
				declaration.Model.IconLarge.EnUS, err = remap(declaration.Model.IconLarge.EnUS)
				if err != nil {
					return nil, err
				}
			}

			if declaration.Model.IconLarge.ZhHans != "" {
				declaration.Model.IconLarge.ZhHans, err = remap(declaration.Model.IconLarge.ZhHans)
				if err != nil {
					return nil, err
				}
			}

			if declaration.Model.IconLarge.JaJp != "" {
				declaration.Model.IconLarge.JaJp, err = remap(declaration.Model.IconLarge.JaJp)
				if err != nil {
					return nil, err
				}
			}

			if declaration.Model.IconLarge.PtBr != "" {
				declaration.Model.IconLarge.PtBr, err = remap(declaration.Model.IconLarge.PtBr)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if declaration.Tool != nil {
		if declaration.Tool.Identity.Icon != "" {
			declaration.Tool.Identity.Icon, err = remap(declaration.Tool.Identity.Icon)
			if err != nil {
				return nil, err
			}
		}
	}

	if declaration.Icon != "" {
		declaration.Icon, err = remap(declaration.Icon)
		if err != nil {
			return nil, err
		}
	}

	return assets_ids, nil
}
