package plugin_entities

import (
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
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

func TestToolParameterShowOn_Validate(t *testing.T) {
	const jsonData = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description"
		},
		"icon": "icon",
		"label": {
			"en_US": "label"
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
					"en_US": "label"
				}
			},
			"description": {
				"human": {
					"en_US": "description"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "mode",
					"type": "select",
					"label": {
						"en_US": "Mode"
					},
					"human_description": {
						"en_US": "Select mode"
					},
					"form": "form",
					"required": true,
					"options": [
						{
							"value": "simple",
							"label": {
								"en_US": "Simple"
							}
						},
						{
							"value": "advanced",
							"label": {
								"en_US": "Advanced"
							}
						}
					]
				},
				{
					"name": "advanced_param",
					"type": "string",
					"label": {
						"en_US": "Advanced Param"
					},
					"human_description": {
						"en_US": "Advanced parameter"
					},
					"form": "form",
					"required": false,
					"show_on": [
						{
							"variable": "mode",
							"value": "advanced"
						}
					]
				}
			]
		}
	]
}
`

	const yamlData = `identity:
  author: author
  name: name
  description:
    en_US: description
  icon: icon
  label:
    en_US: label
  tags: []
credentials_schema: []
tools:
  - identity:
      author: author
      name: tool
      label:
        en_US: label
    description:
      human:
        en_US: description
      llm: description
    parameters:
      - name: mode
        type: select
        label:
          en_US: Mode
        human_description:
          en_US: Select mode
        form: form
        required: true
        options:
          - value: simple
            label:
              en_US: Simple
          - value: advanced
            label:
              en_US: Advanced
      - name: advanced_param
        type: string
        label:
          en_US: Advanced Param
        human_description:
          en_US: Advanced parameter
        form: form
        required: false
        show_on:
          - variable: mode
            value: advanced
`

	jsonDeclaration, jsonErr := UnmarshalToolProviderDeclaration([]byte(jsonData))
	if jsonErr != nil {
		t.Errorf("UnmarshalToolProviderDeclaration() error for JSON = %v", jsonErr)
		return
	}

	if len(jsonDeclaration.Tools[0].Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(jsonDeclaration.Tools[0].Parameters))
		return
	}

	advancedParam := jsonDeclaration.Tools[0].Parameters[1]
	if len(advancedParam.ShowOn) != 1 {
		t.Errorf("Expected 1 show_on condition, got %d", len(advancedParam.ShowOn))
		return
	}

	if advancedParam.ShowOn[0].Variable != "mode" || advancedParam.ShowOn[0].Value != "advanced" {
		t.Errorf("Unexpected show_on values: variable=%s, value=%s", advancedParam.ShowOn[0].Variable, advancedParam.ShowOn[0].Value)
		return
	}

	modeParam := jsonDeclaration.Tools[0].Parameters[0]
	if modeParam.ShowOn == nil {
		t.Errorf("Expected ShowOn to be initialized to empty slice, got nil")
		return
	}

	yamlDeclaration, yamlErr := parser.UnmarshalYamlBytes[ToolProviderDeclaration]([]byte(yamlData))
	if yamlErr != nil {
		t.Errorf("UnmarshalToolProviderDeclaration() error for YAML = %v", yamlErr)
		return
	}

	if len(yamlDeclaration.Tools[0].Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(yamlDeclaration.Tools[0].Parameters))
		return
	}

	yamlAdvancedParam := yamlDeclaration.Tools[0].Parameters[1]
	if len(yamlAdvancedParam.ShowOn) != 1 {
		t.Errorf("Expected 1 show_on condition, got %d", len(yamlAdvancedParam.ShowOn))
		return
	}

	if yamlAdvancedParam.ShowOn[0].Variable != "mode" || yamlAdvancedParam.ShowOn[0].Value != "advanced" {
		t.Errorf("Unexpected show_on values: variable=%s, value=%s", yamlAdvancedParam.ShowOn[0].Variable, yamlAdvancedParam.ShowOn[0].Value)
		return
	}
}

func TestParameterOptionShowOn_Validate(t *testing.T) {
	const jsonData = `
{
	"identity": {
		"author": "author",
		"name": "name",
		"description": {
			"en_US": "description"
		},
		"icon": "icon",
		"label": {
			"en_US": "label"
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
					"en_US": "label"
				}
			},
			"description": {
				"human": {
					"en_US": "description"
				},
				"llm": "description"
			},
			"parameters": [
				{
					"name": "mode",
					"type": "select",
					"label": {
						"en_US": "Mode"
					},
					"human_description": {
						"en_US": "Select mode"
					},
					"form": "form",
					"required": true,
					"options": [
						{
							"value": "simple",
							"label": {
								"en_US": "Simple"
							}
						},
						{
							"value": "advanced",
							"label": {
								"en_US": "Advanced"
							},
							"show_on": [
								{
									"variable": "feature_flag",
									"value": "true"
								}
							]
						}
					]
				}
			]
		}
	]
}
`

	declaration, err := UnmarshalToolProviderDeclaration([]byte(jsonData))
	if err != nil {
		t.Errorf("UnmarshalToolProviderDeclaration() error = %v", err)
		return
	}

	modeParam := declaration.Tools[0].Parameters[0]
	if len(modeParam.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(modeParam.Options))
		return
	}

	advancedOption := modeParam.Options[1]
	if len(advancedOption.ShowOn) != 1 {
		t.Errorf("Expected 1 show_on condition on option, got %d", len(advancedOption.ShowOn))
		return
	}

	if advancedOption.ShowOn[0].Variable != "feature_flag" || advancedOption.ShowOn[0].Value != "true" {
		t.Errorf("Unexpected show_on values on option: variable=%s, value=%s", advancedOption.ShowOn[0].Variable, advancedOption.ShowOn[0].Value)
		return
	}

	simpleOption := modeParam.Options[0]
	if simpleOption.ShowOn == nil {
		t.Errorf("Expected ShowOn to be initialized to empty slice on option, got nil")
		return
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
