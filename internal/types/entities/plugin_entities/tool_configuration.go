package plugin_entities

type ToolIdentity struct {
	Author string     `json:"author"`
	Name   string     `json:"name"`
	Label  I18nObject `json:"label"`
}

type ToolParameterOption struct {
	Value string     `json:"value"`
	Label I18nObject `json:"label"`
}

type ToolParameterType string

const (
	TOOL_PARAMETER_TYPE_STRING       ToolParameterType = "string"
	TOOL_PARAMETER_TYPE_NUMBER       ToolParameterType = "number"
	TOOL_PARAMETER_TYPE_BOOLEAN      ToolParameterType = "boolean"
	TOOL_PARAMETER_TYPE_SELECT       ToolParameterType = "select"
	TOOL_PARAMETER_TYPE_SECRET_INPUT ToolParameterType = "secret_input"
	TOOL_PARAMETER_TYPE_FILE         ToolParameterType = "file"
)

type ToolParameterForm string

const (
	TOOL_PARAMETER_FORM_SCHEMA ToolParameterForm = "schema"
	TOOL_PARAMETER_FORM_FORM   ToolParameterForm = "form"
	TOOL_PARAMETER_FORM_LLM    ToolParameterForm = "llm"
)

type ToolParameter struct {
	Name             string                `json:"name"`
	Label            I18nObject            `json:"label"`
	HumanDescription I18nObject            `json:"human_description"`
	Type             ToolParameterType     `json:"type"`
	Form             ToolParameterForm     `json:"form"`
	LLMDescription   string                `json:"llm_description"`
	Required         bool                  `json:"required"`
	Default          any                   `json:"default"`
	Min              *float64              `json:"min"`
	Max              *float64              `json:"max"`
	Options          []ToolParameterOption `json:"options"`
}

type ToolDescription struct {
	Human I18nObject `json:"human"`
	LLM   string     `json:"llm"`
}

type ToolConfiguration struct {
	Identity    ToolIdentity    `json:"identity"`
	Description ToolDescription `json:"description"`
	Parameters  []ToolParameter `json:"parameters"`
}

type ToolCredentialsOption struct {
	Value string     `json:"value"`
	Label I18nObject `json:"label"`
}

type CredentialType string

const (
	CREDENTIALS_TYPE_SECRET_INPUT CredentialType = "secret_input"
	CREDENTIALS_TYPE_TEXT_INPUT   CredentialType = "text_input"
	CREDENTIALS_TYPE_SELECT       CredentialType = "select"
	CREDENTIALS_TYPE_BOOLEAN      CredentialType = "boolean"
)

type ToolProviderCredential struct {
	Name        string                  `json:"name"`
	Type        CredentialType          `json:"type"`
	Required    bool                    `json:"required"`
	Default     any                     `json:"default"`
	Options     []ToolCredentialsOption `json:"options"`
	Label       I18nObject              `json:"label"`
	Helper      *I18nObject             `json:"helper"`
	URL         *string                 `json:"url"`
	Placeholder *I18nObject             `json:"placeholder"`
}

type ToolLabel string

const (
	TOOL_LABEL_SEARCH        ToolLabel = "search"
	TOOL_LABEL_IMAGE         ToolLabel = "image"
	TOOL_LABEL_VIDEOS        ToolLabel = "videos"
	TOOL_LABEL_WEATHER       ToolLabel = "weather"
	TOOL_LABEL_FINANCE       ToolLabel = "finance"
	TOOL_LABEL_DESIGN        ToolLabel = "design"
	TOOL_LABEL_TRAVEL        ToolLabel = "travel"
	TOOL_LABEL_SOCIAL        ToolLabel = "social"
	TOOL_LABEL_NEWS          ToolLabel = "news"
	TOOL_LABEL_MEDICAL       ToolLabel = "medical"
	TOOL_LABEL_PRODUCTIVITY  ToolLabel = "productivity"
	TOOL_LABEL_EDUCATION     ToolLabel = "education"
	TOOL_LABEL_BUSINESS      ToolLabel = "business"
	TOOL_LABEL_ENTERTAINMENT ToolLabel = "entertainment"
	TOOL_LABEL_UTILITIES     ToolLabel = "utilities"
	TOOL_LABEL_OTHER         ToolLabel = "other"
)

type ToolProviderIdentity struct {
	Author      string      `json:"author"`
	Name        string      `json:"name"`
	Description I18nObject  `json:"description"`
	Icon        []byte      `json:"icon"`
	Label       I18nObject  `json:"label"`
	Tags        []ToolLabel `json:"tags"`
}

type ToolProviderConfiguration struct {
	Identity          ToolProviderIdentity              `json:"identity"`
	CredentialsSchema map[string]ToolProviderCredential `json:"credentials_schema"`
	Tools             []ToolConfiguration               `json:"tools"`
}
