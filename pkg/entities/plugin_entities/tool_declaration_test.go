package plugin_entities

import (
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities/builtin_schema"
	"github.com/stretchr/testify/assert"
)

func TestFullFunctionToolProvider_Validate(t *testing.T) {
	const json_data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": [
			"image",
			"videos"
		]
	},
	"credentials_schema": [
		{
			"name": "api_key",
			"type": "secret-input",
			"required": false,
			"default": "default",
			"label": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"help": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			}
		}
	],
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "parameter",
					"type": "string",
					"label": {
						"en_US": "label",
						"zh_Hans": "标签",
						"pt_BR": "etiqueta"
					},
					"human_description": {
						"en_US": "description",
						"zh_Hans": "描述",
						"pt_BR": "descrição"
					},
					"form": "llm",
					"required": true,
					"default": "default",
					"options": [
						{
							"value": "value",
							"label": {
								"en_US": "label",
								"zh_Hans": "标签",
								"pt_BR": "etiqueta"
							}
						}
					]
				}
			]
		}
	]
}
	`

	const yaml_data = `identity:
  author: author
  name: name
  description:
    en_US: description
    zh_Hans: 描述
    pt_BR: descrição
  icon: icon
  label:
    en_US: label
    zh_Hans: 标签
    pt_BR: etiqueta
  tags:
    - image
    - videos
credentials_schema:
  - name: api_key
    type: secret-input
    required: false
    default: default
    label:
      en_US: API Key
      zh_Hans: API 密钥
      pt_BR: Chave da API
    help:
      en_US: API Key
      zh_Hans: API 密钥
      pt_BR: Chave da API
    url: https://example.com
    placeholder:
      en_US: API Key
      zh_Hans: API 密钥
      pt_BR: Chave da API
tools:
  - identity:
      author: author
      name: tool
      label:
        en_US: label
        zh_Hans: 标签
        pt_BR: etiqueta
    description:
      human:
        en_US: description
        zh_Hans: 描述
        pt_BR: descrição
      llm: description
    parameters:
      - name: parameter
        type: string
        label:
          en_US: label
          zh_Hans: 标签
          pt_BR: etiqueta
        human_description:
          en_US: description
          zh_Hans: 描述
          pt_BR: descrição
        form: llm
        required: true
        default: default
        options:
          - value: value
            label:
              en_US: label
              zh_Hans: 标签
              pt_BR: etiqueta
    `

	jsonDeclaration, jsonErr := parser.UnmarshalJsonBytes[ToolProviderDeclaration]([]byte(json_data))
	if jsonErr != nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error for JSON = %v", jsonErr)
		return
	}

	if len(jsonDeclaration.CredentialsSchema) != 1 {
		t.Errorf("UnmarshalToolProviderConfiguration() error for JSON: incorrect CredentialsSchema length")
		return
	}

	yamlDeclaration, yamlErr := parser.UnmarshalYamlBytes[ToolProviderDeclaration]([]byte(yaml_data))
	if yamlErr != nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error for YAML = %v", yamlErr)
		return
	}

	if len(yamlDeclaration.CredentialsSchema) != 1 {
		t.Errorf("UnmarshalToolProviderConfiguration() error for YAML: incorrect CredentialsSchema length")
		return
	}
}

func TestToolProviderWithMapCredentials_Validate(t *testing.T) {
	const json_data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": [
			"image",
			"videos"
		]
	},
	"credentials_schema": {
		"api_key": {
			"type": "secret-input",
			"required": false,
			"default": "default",
			"label": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"help": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			}
		}
	},
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "parameter",
					"type": "string",
					"label": {
						"en_US": "label",
						"zh_Hans": "标签",
						"pt_BR": "etiqueta"
					},
					"human_description": {
						"en_US": "description",
						"zh_Hans": "描述",
						"pt_BR": "descrição"
					},
					"form": "llm",
					"required": true,
					"default": "default",
					"options": [
						{
							"value": "value",
							"label": {
								"en_US": "label",
								"zh_Hans": "标签",
								"pt_BR": "etiqueta"
							}
						}
					]
				}
			]
		}
	]
}
	`

	const yaml_data = `identity:
  author: author
  name: name
  description:
    en_US: description
    zh_Hans: 描述
    pt_BR: descrição
  icon: icon
  label:
    en_US: label
    zh_Hans: 标签
    pt_BR: etiqueta
  tags:
    - image
    - videos
credentials_schema:
  api_key:
    type: secret-input
    required: false
    default: default
    label:
      en_US: API Key
      zh_Hans: API 密钥
      pt_BR: Chave da API
    help:
      en_US: API Key
      zh_Hans: API 密钥
      pt_BR: Chave da API
    url: https://example.com
    placeholder:
      en_US: API Key
      zh_Hans: API 密钥
      pt_BR: Chave da API
tools:
  - identity:
      author: author
      name: tool
      label:
        en_US: label
        zh_Hans: 标签
        pt_BR: etiqueta
    description:
      human:
        en_US: description
        zh_Hans: 描述
        pt_BR: descrição
      llm: description
    parameters:
      - name: parameter
        type: string
        label:
          en_US: label
          zh_Hans: 标签
          pt_BR: etiqueta
        human_description:
          en_US: description
          zh_Hans: 描述
          pt_BR: descrição
        form: llm
        required: true
        default: default
        options:
          - value: value
            label:
              en_US: label
              zh_Hans: 标签
              pt_BR: etiqueta
    `

	jsonDeclaration, jsonErr := parser.UnmarshalJsonBytes[ToolProviderDeclaration]([]byte(json_data))
	if jsonErr != nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error for JSON = %v", jsonErr)
		return
	}

	if len(jsonDeclaration.CredentialsSchema) != 1 {
		t.Errorf("UnmarshalToolProviderConfiguration() error for JSON: incorrect CredentialsSchema length")
		return
	}

	if len(jsonDeclaration.Tools) != 1 {
		t.Errorf("UnmarshalToolProviderConfiguration() error for JSON: incorrect Tools length")
		return
	}

	yamlDeclaration, yamlErr := parser.UnmarshalYamlBytes[ToolProviderDeclaration]([]byte(yaml_data))
	if yamlErr != nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error for YAML = %v", yamlErr)
		return
	}

	if len(yamlDeclaration.CredentialsSchema) != 1 {
		t.Errorf("UnmarshalToolProviderConfiguration() error for YAML: incorrect CredentialsSchema length")
		return
	}

	if len(yamlDeclaration.Tools) != 1 {
		t.Errorf("UnmarshalToolProviderConfiguration() error for YAML: incorrect Tools length")
		return
	}
}

func TestWithoutAuthorToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": [
			"image",
			"videos"
		]
	},
	"credentials_schema": [
		{
			"name": "api_key",
			"type": "secret-input",
			"required": false,
			"default": "default",
			"label": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"help": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			}
		}
	]
},
	"tools": [
	
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWithoutNameToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": [
			"image",
			"videos"
		]
	},
	"credentials_schema": [
		{
			"name": "api_key",
			"type": "secret-input",
			"type": "secret-input",
			"required": false,
			"default": "default",
			"label": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"help": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			}
		}
	],
	"tools": [
	
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWithoutDescriptionToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": [
			"image",
			"videos"
		]
	},
	"credentials_schema": [
		{
			"name": "api_key",
			"type": "secret-input",
			"required": false,
			"default": "default",
			"label": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"help": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			}
		}
	],
	"tools": [
	
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWrongCredentialTypeToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": [
			"image",
			"videos"
		]
	},
	"credentials_schema": [
		{
			"name": "api_key",
			"type": "wrong",
			"required": false,
			"default": "default",
			"label": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"help": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			}
		}
	],
	"tools": [
	
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWrongIdentityTagsToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": [
			"wrong",
			"videos"
		]
	},
	"credentials_schema": [
		{
			"name": "api_key",
			"type": "secret-input",
			"required": false,
			"default": "default",
			"label": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"help": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "API Key",
				"zh_Hans": "API 密钥",
				"pt_BR": "Chave da API"
			}
		}
	],
	"tools": [
	
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWrongToolParameterTypeToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": []
	},
	"credentials_schema": [],
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "parameter",
					"type": "wrong",
					"label": {
						"en_US": "label",
						"zh_Hans": "标签",
						"pt_BR": "etiqueta"
					},
					"human_description": {
						"en_US": "description",
						"zh_Hans": "描述",
						"pt_BR": "descrição"
					},
					"form": "llm",
					"required": true,
					"default": "default",
					"options": []
				}
			]
		}
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWrongToolParameterFormToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": []
	},
	"credentials_schema": [],
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "parameter",
					"type": "string",
					"label": {
						"en_US": "label",
						"zh_Hans": "标签",
						"pt_BR": "etiqueta"
					},
					"human_description": {
						"en_US": "description",
						"zh_Hans": "描述",
						"pt_BR": "descrição"
					},
					"form": "wrong",
					"required": true,
					"default": "default",
					"options": []
				}
			]
		}
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestJSONSchemaTypeToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": []
	},
	"credentials_schema": [],
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"output_schema": {
				"type": "object",
				"properties": {
					"name": {
						"type": "string"
					}
				}
			}
		}
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err != nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWrongJSONSchemaToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": []
	},
	"credentials_schema": [],
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"output_schema": {
				"type": "object",
				"properties": {
					"name": {
						"type": "aaa"
					}
				}
			}
		}
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestWrongAppSelectorScopeToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": []
	},
	"credentials_schema": [
		{
			"name": "api_key",
			"type": "app-selector",
			"scope": "wrong",
			"required": false,
			"default": null,
			"label": {
				"en_US": "app-selector",
				"zh_Hans": "app-selector",
				"pt_BR": "app-selector"
			},
			"help": {
				"en_US": "app-selector",
				"zh_Hans": "app-selector",
				"pt_BR": "app-selector"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "app-selector",
				"zh_Hans": "app-selector",
				"pt_BR": "app-selector"
			}
		}
	],
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "parameter-app-selector",
					"label": {
						"en_US": "label",
						"zh_Hans": "标签",
						"pt_BR": "etiqueta"
					},
					"human_description": {
						"en_US": "description",
						"zh_Hans": "描述",
						"pt_BR": "descrição"
					},
					"type": "app-selector",
					"form": "llm",
					"scope": "wrong",
					"required": true,
					"default": "default",
					"options": []
				}
			]
		}
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err == nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}

	str := err.Error()
	if !strings.Contains(str, "is_scope") {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}

	if !strings.Contains(str, "ToolProviderDeclaration.Tools[0].Parameters[0].Scope") {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestAppSelectorScopeToolProvider_Validate(t *testing.T) {
	const data = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description",
			"zh_Hans": "描述",
			"pt_BR": "descrição"
		},
		"icon": "icon",
		"label": {
			"en_US": "label",
			"zh_Hans": "标签",
			"pt_BR": "etiqueta"
		},
		"tags": []
	},
	"credentials_schema": [
		{
			"name": "app-selector",
			"type": "app-selector",
			"scope": "all",
			"required": false,
			"default": null,
			"label": {
				"en_US": "app-selector",
				"zh_Hans": "app-selector",
				"pt_BR": "app-selector"
			},
			"help": {
				"en_US": "app-selector",
				"zh_Hans": "app-selector",
				"pt_BR": "app-selector"
			},
			"url": "https://example.com",
			"placeholder": {
				"en_US": "app-selector",
				"zh_Hans": "app-selector",
				"pt_BR": "app-selector"
			}
		}
	],
	"tools": [
		{
			"identity": {
				"author": "author",
				"name": "tool",
				"label": {
					"en_US": "label",
					"zh_Hans": "标签",
					"pt_BR": "etiqueta"
				}
			},
			"description": {
				"human": {
					"en_US": "description",
					"zh_Hans": "描述",
					"pt_BR": "descrição"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "parameter-app-selector",
					"label": {
						"en_US": "label",
						"zh_Hans": "标签",
						"pt_BR": "etiqueta"
					},
					"human_description": {
						"en_US": "description",
						"zh_Hans": "描述",
						"pt_BR": "descrição"
					},
					"type": "app-selector",
					"form": "llm",
					"scope": "all",
					"required": true,
					"default": "default",
					"options": []
				}
			]
		}
	]
}
	`

	_, err := UnmarshalToolProviderDeclaration([]byte(data))
	if err != nil {
		t.Errorf("UnmarshalToolProviderConfiguration() error = %v, wantErr %v", err, true)
		return
	}
}

func TestParameterScope_Validate(t *testing.T) {
	config := ToolParameter{
		Name:     "test",
		Type:     TOOL_PARAMETER_TYPE_MODEL_SELECTOR,
		Scope:    parser.ToPtr("llm& document&tool-call"),
		Required: true,
		Label: I18nObject{
			ZhHans: "模型",
			EnUS:   "Model",
		},
		HumanDescription: I18nObject{
			ZhHans: "请选择模型",
			EnUS:   "Please select a model",
		},
		LLMDescription: "please select a model",
		Form:           TOOL_PARAMETER_FORM_FORM,
	}

	data := parser.MarshalJsonBytes(config)

	if _, err := parser.UnmarshalJsonBytes[ToolParameter](data); err != nil {
		t.Errorf("ParameterScope_Validate() error = %v", err)
	}
}

func TestInvalidJSONSchemaToolProvider_Validate(t *testing.T) {
	type Test struct {
		Text ToolOutputSchema `json:"text" validate:"json_schema"`
	}

	data := parser.MarshalJsonBytes(Test{
		Text: map[string]any{
			"text": "text",
		},
	})

	if _, err := parser.UnmarshalJsonBytes[Test](data); err == nil {
		t.Errorf("TestInvalidJSONSchemaToolProvider_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(Test{
		Text: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{
					"type": "object",
				},
			},
		},
	})

	if _, err := parser.UnmarshalJsonBytes[Test](data); err == nil {
		t.Errorf("TestInvalidJSONSchemaToolProvider_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(Test{
		Text: map[string]any{
			"type": "array",
			"properties": map[string]any{
				"a": map[string]any{
					"type": "object",
				},
			},
		},
	})

	if _, err := parser.UnmarshalJsonBytes[Test](data); err == nil {
		t.Errorf("TestInvalidJSONSchemaToolProvider_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(Test{
		Text: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"json": map[string]any{
					"type": "object",
				},
			},
		},
	})

	if _, err := parser.UnmarshalJsonBytes[Test](data); err == nil {
		t.Errorf("TestInvalidJSONSchemaToolProvider_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(Test{
		Text: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"files": map[string]any{
					"type": "object",
				},
			},
		},
	})

	if _, err := parser.UnmarshalJsonBytes[Test](data); err == nil {
		t.Errorf("TestInvalidJSONSchemaToolProvider_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(Test{
		Text: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"aaa": map[string]any{
					"type": "object",
				},
			},
		},
	})

	if _, err := parser.UnmarshalJsonBytes[Test](data); err != nil {
		t.Errorf("TestInvalidJSONSchemaToolProvider_Validate() error = %v", err)
	}
}

func TestToolName_Validate(t *testing.T) {
	data := parser.MarshalJsonBytes(ToolProviderIdentity{
		Author: "author",
		Name:   "tool-name",
		Description: I18nObject{
			EnUS:   "description",
			ZhHans: "描述",
		},
		Icon: "icon",
		Label: I18nObject{
			EnUS:   "label",
			ZhHans: "标签",
		},
	})

	if _, err := parser.UnmarshalJsonBytes[ToolProviderIdentity](data); err != nil {
		t.Errorf("TestToolName_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(ToolProviderIdentity{
		Author: "author",
		Name:   "tool AA",
		Label: I18nObject{
			EnUS:   "label",
			ZhHans: "标签",
		},
	})

	if _, err := parser.UnmarshalJsonBytes[ToolProviderIdentity](data); err == nil {
		t.Errorf("TestToolName_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(ToolIdentity{
		Author: "author",
		Name:   "tool-name-123",
		Label: I18nObject{
			EnUS:   "label",
			ZhHans: "标签",
		},
	})

	if _, err := parser.UnmarshalJsonBytes[ToolIdentity](data); err != nil {
		t.Errorf("TestToolName_Validate() error = %v", err)
	}

	data = parser.MarshalJsonBytes(ToolIdentity{
		Author: "author",
		Name:   "tool name-123",
		Label: I18nObject{
			EnUS:   "label",
			ZhHans: "标签",
		},
	})

	if _, err := parser.UnmarshalJsonBytes[ToolIdentity](data); err == nil {
		t.Errorf("TestToolName_Validate() error = %v", err)
	}
}

// Tests for schema reference processing in tools
func TestProcessToolYAML(t *testing.T) {
	t.Run("process output_schema with refs", func(t *testing.T) {
		yamlData := map[string]any{
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"$ref": "#/$defs/file",
					},
				},
			},
		}

		processed, err := ProcessToolYAML(yamlData)
		assert.NoError(t, err)

		outputSchema := processed["output_schema"].(map[string]any)
		result := outputSchema["properties"].(map[string]any)["result"].(map[string]any)
		assert.Equal(t, "object", result["type"])
		assert.Contains(t, result, "properties")
		
		// Verify the file schema was resolved correctly
		props := result["properties"].(map[string]any)
		assert.Contains(t, props, "name")
		assert.Contains(t, props, "dify_builtin_type")
	})

	t.Run("process mixed user and builtin definitions", func(t *testing.T) {
		yamlData := map[string]any{
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file": map[string]any{
						"$ref": "#/$defs/file",
					},
					"custom": map[string]any{
						"$ref": "#/$defs/my_type",
					},
				},
			},
			"definitions": map[string]any{
				"my_type": map[string]any{
					"type": "string",
					"description": "Custom type",
				},
			},
		}

		processed, err := ProcessToolYAML(yamlData)
		assert.NoError(t, err)

		// Definitions should be removed after processing
		assert.NotContains(t, processed, "definitions")

		outputSchema := processed["output_schema"].(map[string]any)
		props := outputSchema["properties"].(map[string]any)

		// Check builtin file type was resolved
		fileRef := props["file"].(map[string]any)
		assert.Equal(t, "object", fileRef["type"])
		assert.Contains(t, fileRef, "properties")

		// Check custom type was resolved
		customRef := props["custom"].(map[string]any)
		assert.Equal(t, "string", customRef["type"])
		assert.Equal(t, "Custom type", customRef["description"])
	})

	t.Run("process complex nested refs", func(t *testing.T) {
		yamlData := map[string]any{
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"files": map[string]any{
						"type": "array",
						"items": map[string]any{
							"$ref": "#/$defs/file",
						},
					},
					"crawl_result": map[string]any{
						"$ref": "#/$defs/website_crawl",
					},
				},
			},
		}

		processed, err := ProcessToolYAML(yamlData)
		assert.NoError(t, err)

		outputSchema := processed["output_schema"].(map[string]any)
		props := outputSchema["properties"].(map[string]any)

		// Check array of files
		filesArray := props["files"].(map[string]any)
		fileItems := filesArray["items"].(map[string]any)
		assert.Equal(t, "object", fileItems["type"])
		assert.Contains(t, fileItems, "properties")

		// Check website crawl result
		crawlResult := props["crawl_result"].(map[string]any)
		assert.Equal(t, "object", crawlResult["type"])
		crawlProps := crawlResult["properties"].(map[string]any)
		assert.Contains(t, crawlProps, "source_url")
		assert.Contains(t, crawlProps, "content")
	})

	t.Run("no output_schema should remain unchanged", func(t *testing.T) {
		yamlData := map[string]any{
			"identity": map[string]any{
				"name": "test_tool",
			},
			"description": map[string]any{
				"human": map[string]any{
					"en_US": "Test tool",
				},
				"llm": "A test tool",
			},
		}

		processed, err := ProcessToolYAML(yamlData)
		assert.NoError(t, err)

		// Data should remain unchanged
		assert.Equal(t, yamlData, processed)
	})

	t.Run("error on non-existent ref", func(t *testing.T) {
		yamlData := map[string]any{
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"$ref": "#/$defs/non_existent_type",
					},
				},
			},
		}

		_, err := ProcessToolYAML(yamlData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestToolSchemaRefsIntegration(t *testing.T) {
	t.Run("test with various builtin types", func(t *testing.T) {
		builtinTypes := []string{
			"file",
			"website_crawl", 
			"online_document",
			"general_structure_chunk",
			"parent_child_structure_chunk",
			"qa_structure_chunk",
		}

		for _, typeName := range builtinTypes {
			t.Run("builtin_type_"+typeName, func(t *testing.T) {
				yamlData := map[string]any{
					"output_schema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"result": map[string]any{
								"$ref": "#/$defs/" + typeName,
							},
						},
					},
				}

				processed, err := ProcessToolYAML(yamlData)
				assert.NoError(t, err)

				outputSchema := processed["output_schema"].(map[string]any)
				result := outputSchema["properties"].(map[string]any)["result"].(map[string]any)
				assert.Equal(t, "object", result["type"])
				
				// All builtin types should have dify_builtin_type
				props := result["properties"].(map[string]any)
				assert.Contains(t, props, "dify_builtin_type")
				
				builtinType := props["dify_builtin_type"].(map[string]any)
				assert.Contains(t, builtinType, "enum")
			})
		}
	})

	t.Run("user definitions override builtin", func(t *testing.T) {
		yamlData := map[string]any{
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"$ref": "#/$defs/file",
					},
				},
			},
			"definitions": map[string]any{
				"file": map[string]any{
					"type": "string",
					"description": "Custom file type",
				},
			},
		}

		processed, err := ProcessToolYAML(yamlData)
		assert.NoError(t, err)

		outputSchema := processed["output_schema"].(map[string]any)
		result := outputSchema["properties"].(map[string]any)["result"].(map[string]any)
		
		// Should use user definition, not builtin
		assert.Equal(t, "string", result["type"])
		assert.Equal(t, "Custom file type", result["description"])
		assert.NotContains(t, result, "properties") // builtin file has properties
	})
}

func TestBuiltinSchemaAccessFromTool(t *testing.T) {
	t.Run("builtin definitions should be accessible", func(t *testing.T) {
		// Test that we can access builtin definitions from tool context
		assert.Contains(t, builtin_schema.BuiltinDefinitions, "file")
		assert.Contains(t, builtin_schema.BuiltinDefinitions, "website_crawl")
		
		fileDef := builtin_schema.BuiltinDefinitions["file"].(map[string]any)
		assert.Equal(t, "object", fileDef["type"])
		assert.Contains(t, fileDef, "properties")
	})

	t.Run("resolve schema refs should work directly", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"upload": map[string]any{
					"$ref": "#/$defs/file",
				},
			},
		}

		resolved, err := builtin_schema.ResolveSchemaRefs(schema, builtin_schema.BuiltinDefinitions)
		assert.NoError(t, err)

		props := resolved.(map[string]any)["properties"].(map[string]any)
		upload := props["upload"].(map[string]any)
		assert.Equal(t, "object", upload["type"])
	})
}
