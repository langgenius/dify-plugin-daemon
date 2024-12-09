package media_manager

import (
	"errors"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
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
		if declaration.Model.IconSmall != nil {
			if declaration.Model.IconSmall.EnUS != "" {
				declaration.Model.IconSmall.EnUS, err = remap(declaration.Model.IconSmall.EnUS)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon small en_US"))
				}
			}

			if declaration.Model.IconSmall.ZhHans != "" {
				declaration.Model.IconSmall.ZhHans, err = remap(declaration.Model.IconSmall.ZhHans)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon small zh_Hans"))
				}
			}

			if declaration.Model.IconSmall.JaJp != "" {
				declaration.Model.IconSmall.JaJp, err = remap(declaration.Model.IconSmall.JaJp)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon small ja_JP"))
				}
			}

			if declaration.Model.IconSmall.PtBr != "" {
				declaration.Model.IconSmall.PtBr, err = remap(declaration.Model.IconSmall.PtBr)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon small pt_BR"))
				}
			}
		}

		if declaration.Model.IconLarge != nil {
			if declaration.Model.IconLarge.EnUS != "" {
				declaration.Model.IconLarge.EnUS, err = remap(declaration.Model.IconLarge.EnUS)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon large en_US"))
				}
			}

			if declaration.Model.IconLarge.ZhHans != "" {
				declaration.Model.IconLarge.ZhHans, err = remap(declaration.Model.IconLarge.ZhHans)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon large zh_Hans"))
				}
			}

			if declaration.Model.IconLarge.JaJp != "" {
				declaration.Model.IconLarge.JaJp, err = remap(declaration.Model.IconLarge.JaJp)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon large ja_JP"))
				}
			}

			if declaration.Model.IconLarge.PtBr != "" {
				declaration.Model.IconLarge.PtBr, err = remap(declaration.Model.IconLarge.PtBr)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to remap model icon large pt_BR"))
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
	}

	if declaration.Agent != nil {
		if declaration.Agent.Identity.Icon != "" {
			declaration.Agent.Identity.Icon, err = remap(declaration.Agent.Identity.Icon)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to remap agent icon"))
			}
		}
	}

	if declaration.Icon != "" {
		declaration.Icon, err = remap(declaration.Icon)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("failed to remap plugin icon"))
		}
	}

	return assetsIds, nil
}
