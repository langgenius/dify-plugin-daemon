package plugin_entities

import (
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
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

// TestToolOutputSchemaRefExpansion tests that $ref references are automatically expanded
func TestToolOutputSchemaRefExpansion(t *testing.T) {
	// Test YAML with built-in $ref
	yamlData := `
identity:
  author: test
  name: test_tool
  label:
    en_US: Test Tool
description:
  human:
    en_US: Test tool
  llm: Test tool
output_schema:
  type: object
  properties:
    file_result:
      $ref: "#/$defs/file"
    chunk_result:
      $ref: "#/$defs/general_structure_chunk"
`

	toolDecl, err := parser.UnmarshalYamlBytes[ToolDeclaration]([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML with $ref: %v", err)
	}

	// Check if file_result was expanded
	if toolDecl.OutputSchema == nil {
		t.Fatal("OutputSchema should not be nil")
	}

	properties, ok := toolDecl.OutputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("OutputSchema.properties should be a map")
	}

	// Check file_result expansion
	fileResult, ok := properties["file_result"].(map[string]any)
	if !ok {
		t.Fatal("file_result should be expanded to a map")
	}

	if fileResult["type"] != "object" {
		t.Errorf("file_result.type should be 'object', got %v", fileResult["type"])
	}

	fileProps, ok := fileResult["properties"].(map[string]any)
	if !ok {
		t.Fatal("file_result.properties should exist")
	}

	// Check for expected file properties
	expectedFileProps := []string{"name", "size", "file_type", "extension", "mime_type", "transfer_method", "url", "related_id"}
	for _, prop := range expectedFileProps {
		if _, exists := fileProps[prop]; !exists {
			t.Errorf("file_result.properties.%s should exist", prop)
		}
	}

	// Check chunk_result expansion
	chunkResult, ok := properties["chunk_result"].(map[string]any)
	if !ok {
		t.Fatal("chunk_result should be expanded to a map")
	}

	if chunkResult["type"] != "object" {
		t.Errorf("chunk_result.type should be 'object', got %v", chunkResult["type"])
	}
}

// TestToolOutputSchemaCustomDefinitionsNotSupported tests that custom definitions are not supported
func TestToolOutputSchemaCustomDefinitionsNotSupported(t *testing.T) {
	jsonData := `{
		"identity": {
			"author": "test",
			"name": "test_tool",
			"label": {"en_US": "Test Tool"}
		},
		"description": {
			"human": {"en_US": "Test tool"},
			"llm": "Test tool"
		},
		"output_schema": {
			"type": "object",
			"properties": {
				"custom_result": {"$ref": "#/$defs/custom_type"},
				"file_result": {"$ref": "#/$defs/file"}
			},
			"definitions": {
				"custom_type": {
					"type": "object",
					"properties": {
						"custom_field": {"type": "string"}
					}
				}
			}
		}
	}`

	_, err := parser.UnmarshalJsonBytes[ToolDeclaration]([]byte(jsonData))
	if err == nil {
		t.Error("Should fail with custom definitions that are not supported")
	}
	
	// Should fail because custom_type is not a built-in definition
	if !strings.Contains(err.Error(), "custom_type") {
		t.Errorf("Error should mention the unsupported custom reference, got: %v", err)
	}
}

// TestToolOutputSchemaNestedRefs tests nested $ref references
func TestToolOutputSchemaNestedRefs(t *testing.T) {
	yamlData := `
identity:
  author: test
  name: test_tool
  label:
    en_US: Test Tool
description:
  human:
    en_US: Test tool
  llm: Test tool
output_schema:
  type: object
  properties:
    nested:
      type: object
      properties:
        file_in_nested:
          $ref: "#/$defs/file"
    direct_file:
      $ref: "#/$defs/file"
`

	toolDecl, err := parser.UnmarshalYamlBytes[ToolDeclaration]([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML with nested $ref: %v", err)
	}

	properties := toolDecl.OutputSchema["properties"].(map[string]any)
	
	// Check nested structure
	nested := properties["nested"].(map[string]any)
	nestedProps := nested["properties"].(map[string]any)
	fileInNested := nestedProps["file_in_nested"].(map[string]any)
	
	if fileInNested["type"] != "object" {
		t.Errorf("file_in_nested should be expanded to object type")
	}

	// Check direct file reference
	directFile := properties["direct_file"].(map[string]any)
	if directFile["type"] != "object" {
		t.Errorf("direct_file should be expanded to object type")
	}
}

// TestToolOutputSchemaInvalidRef tests handling of invalid $ref
func TestToolOutputSchemaInvalidRef(t *testing.T) {
	jsonData := `{
		"identity": {
			"author": "test",
			"name": "test_tool",
			"label": {"en_US": "Test Tool"}
		},
		"description": {
			"human": {"en_US": "Test tool"},
			"llm": "Test tool"
		},
		"output_schema": {
			"type": "object",
			"properties": {
				"invalid_ref": {"$ref": "#/$defs/non_existent"}
			}
		}
	}`

	_, err := parser.UnmarshalJsonBytes[ToolDeclaration]([]byte(jsonData))
	if err == nil {
		t.Error("Should fail with non-existent $ref")
	}
	if !strings.Contains(err.Error(), "non_existent") {
		t.Errorf("Error should mention the non-existent reference, got: %v", err)
	}
}
