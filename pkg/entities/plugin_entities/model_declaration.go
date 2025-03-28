package plugin_entities

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

type ModelType string

const (
	MODEL_TYPE_LLM            ModelType = "llm"
	MODEL_TYPE_TEXT_EMBEDDING ModelType = "text-embedding"
	MODEL_TYPE_RERANKING      ModelType = "rerank"
	MODEL_TYPE_SPEECH2TEXT    ModelType = "speech2text"
	MODEL_TYPE_MODERATION     ModelType = "moderation"
	MODEL_TYPE_TTS            ModelType = "tts"
	MODEL_TYPE_TEXT2IMG       ModelType = "text2img"
)

func isModelType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(MODEL_TYPE_LLM),
		string(MODEL_TYPE_TEXT_EMBEDDING),
		string(MODEL_TYPE_RERANKING),
		string(MODEL_TYPE_SPEECH2TEXT),
		string(MODEL_TYPE_MODERATION),
		string(MODEL_TYPE_TTS),
		string(MODEL_TYPE_TEXT2IMG):
		return true
	}
	return false
}

type ModelProviderConfigurateMethod string

const (
	CONFIGURATE_METHOD_PREDEFINED_MODEL   ModelProviderConfigurateMethod = "predefined-model"
	CONFIGURATE_METHOD_CUSTOMIZABLE_MODEL ModelProviderConfigurateMethod = "customizable-model"
)

func isModelProviderConfigurateMethod(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(CONFIGURATE_METHOD_PREDEFINED_MODEL),
		string(CONFIGURATE_METHOD_CUSTOMIZABLE_MODEL):
		return true
	}
	return false
}

type ModelParameterType string

const (
	PARAMETER_TYPE_FLOAT   ModelParameterType = "float"
	PARAMETER_TYPE_INT     ModelParameterType = "int"
	PARAMETER_TYPE_STRING  ModelParameterType = "string"
	PARAMETER_TYPE_BOOLEAN ModelParameterType = "boolean"
	PARAMETER_TYPE_TEXT    ModelParameterType = "text"
)

func isModelParameterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PARAMETER_TYPE_FLOAT),
		string(PARAMETER_TYPE_INT),
		string(PARAMETER_TYPE_STRING),
		string(PARAMETER_TYPE_BOOLEAN),
		string(PARAMETER_TYPE_TEXT):
		return true
	}
	return false
}

type ModelParameterRule struct {
	Name        string              `json:"name" yaml:"name" validate:"required,lt=256"`
	UseTemplate *string             `json:"use_template" yaml:"use_template" validate:"omitempty,lt=256"`
	Label       *I18nObject         `json:"label" yaml:"label" validate:"omitempty"`
	Type        *ModelParameterType `json:"type" yaml:"type" validate:"omitempty,model_parameter_type"`
	Help        *I18nObject         `json:"help" yaml:"help" validate:"omitempty"`
	Required    bool                `json:"required" yaml:"required"`
	Default     *any                `json:"default" yaml:"default" validate:"omitempty,is_basic_type"`
	Min         *float64            `json:"min" yaml:"min" validate:"omitempty"`
	Max         *float64            `json:"max" yaml:"max" validate:"omitempty"`
	Precision   *int                `json:"precision" yaml:"precision" validate:"omitempty"`
	Options     []string            `json:"options" yaml:"options" validate:"omitempty,dive,lt=256"`
}

type DefaultParameterName string

const (
	TEMPERATURE       DefaultParameterName = "temperature"
	TOP_P             DefaultParameterName = "top_p"
	TOP_K             DefaultParameterName = "top_k"
	PRESENCE_PENALTY  DefaultParameterName = "presence_penalty"
	FREQUENCY_PENALTY DefaultParameterName = "frequency_penalty"
	MAX_TOKENS        DefaultParameterName = "max_tokens"
	RESPONSE_FORMAT   DefaultParameterName = "response_format"
	JSON_SCHEMA       DefaultParameterName = "json_schema"
)

var PARAMETER_RULE_TEMPLATE = map[DefaultParameterName]ModelParameterRule{
	TEMPERATURE: {
		Label: &I18nObject{
			EnUS:   "Temperature",
			ZhHans: "温度",
			JaJp:   "温度",
			PtBr:   "Temperatura",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_FLOAT),
		Help: &I18nObject{
			EnUS:   "Controls randomness. Lower temperature results in less random completions. As the temperature approaches zero, the model will become deterministic and repetitive. Higher temperature results in more random completions.",
			ZhHans: "温度控制随机性。较低的温度会导致较少的随机完成。随着温度接近零，模型将变得确定性和重复性。较高的温度会导致更多的随机完成。",
			JaJp:   "温度はランダム性を制御します。温度が低いほどランダムな完成が少なくなります。温度がゼロに近づくと、モデルは決定論的で繰り返しになります。温度が高いほどランダムな完成が多くなります。",
			PtBr:   "A temperatura controla a aleatoriedade. Menores temperaturas resultam em menos conclusões aleatórias. À medida que a temperatura se aproxima de zero, o modelo se tornará determinístico e repetitivo. Temperaturas mais altas resultam em mais conclusões aleatórias.",
		},
		Required:  false,
		Default:   parser.ToPtr(any(0.0)),
		Min:       parser.ToPtr(0.0),
		Max:       parser.ToPtr(1.0),
		Precision: parser.ToPtr(2),
	},
	TOP_P: {
		Label: &I18nObject{
			EnUS:   "Top P",
			ZhHans: "Top P",
			JaJp:   "Top P",
			PtBr:   "Top P",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_FLOAT),
		Help: &I18nObject{
			EnUS:   "Controls diversity via nucleus sampling: 0.5 means half of all likelihood-weighted options are considered.",
			ZhHans: "通过核心采样控制多样性：0.5表示考虑了一半的所有可能性加权选项。",
			JaJp:   "核サンプリングを通じて多様性を制御します：0.5は、すべての可能性加权オプションの半分を考慮します。",
			PtBr:   "Controla a diversidade via amostragem de núcleo: 0.5 significa que metade das opções com maior probabilidade são consideradas.",
		},
		Required:  false,
		Default:   parser.ToPtr(any(1.0)),
		Min:       parser.ToPtr(0.0),
		Max:       parser.ToPtr(1.0),
		Precision: parser.ToPtr(2),
	},
	TOP_K: {
		Label: &I18nObject{
			EnUS:   "Top K",
			ZhHans: "Top K",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_INT),
		Help: &I18nObject{
			EnUS:   "Limits the number of tokens to consider for each step by keeping only the k most likely tokens.",
			ZhHans: "通过只保留每一步中最可能的 k 个标记来限制要考虑的标记数量。",
		},
		Required:  false,
		Default:   parser.ToPtr(any(50)),
		Min:       parser.ToPtr(1.0),
		Max:       parser.ToPtr(100.0),
		Precision: parser.ToPtr(0),
	},
	PRESENCE_PENALTY: {
		Label: &I18nObject{
			EnUS:   "Presence Penalty",
			ZhHans: "存在惩罚",
			JaJp:   "存在ペナルティ",
			PtBr:   "Penalidade de presença",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_FLOAT),
		Help: &I18nObject{
			EnUS:   "Applies a penalty to the log-probability of tokens already in the text.",
			ZhHans: "对文本中已有的标记的对数概率施加惩罚。",
			JaJp:   "テキストに既に存在するトークンの対数確率にペナルティを適用します。",
			PtBr:   "Aplica uma penalidade à probabilidade logarítmica de tokens já presentes no texto.",
		},
		Required:  false,
		Default:   parser.ToPtr(any(0.0)),
		Min:       parser.ToPtr(0.0),
		Max:       parser.ToPtr(1.0),
		Precision: parser.ToPtr(2),
	},
	FREQUENCY_PENALTY: {
		Label: &I18nObject{
			EnUS:   "Frequency Penalty",
			ZhHans: "频率惩罚",
			JaJp:   "頻度ペナルティ",
			PtBr:   "Penalidade de frequência",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_FLOAT),
		Help: &I18nObject{
			EnUS:   "Applies a penalty to the log-probability of tokens that appear in the text.",
			ZhHans: "对文本中出现的标记的对数概率施加惩罚。",
			JaJp:   "テキストに出現するトークンの対数確率にペナルティを適用します。",
			PtBr:   "Aplica uma penalidade à probabilidade logarítmica de tokens que aparecem no texto.",
		},
		Required:  false,
		Default:   parser.ToPtr(any(0.0)),
		Min:       parser.ToPtr(0.0),
		Max:       parser.ToPtr(1.0),
		Precision: parser.ToPtr(2),
	},
	MAX_TOKENS: {
		Label: &I18nObject{
			EnUS:   "Max Tokens",
			ZhHans: "最大标记",
			JaJp:   "最大トークン",
			PtBr:   "Máximo de tokens",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_INT),
		Help: &I18nObject{
			EnUS:   "Specifies the upper limit on the length of generated results. If the generated results are truncated, you can increase this parameter.",
			ZhHans: "指定生成结果长度的上限。如果生成结果截断，可以调大该参数。",
			JaJp:   "生成結果の長さの上限を指定します。生成結果が切り捨てられた場合は、このパラメータを大きくすることができます。",
			PtBr:   "Especifica o limite superior para o comprimento dos resultados gerados. Se os resultados gerados forem truncados, você pode aumentar este parâmetro.",
		},
		Required:  false,
		Default:   parser.ToPtr(any(64)),
		Min:       parser.ToPtr(1.0),
		Max:       parser.ToPtr(2048.0),
		Precision: parser.ToPtr(0),
	},
	RESPONSE_FORMAT: {
		Label: &I18nObject{
			EnUS:   "Response Format",
			ZhHans: "回复格式",
			JaJp:   "応答形式",
			PtBr:   "Formato de resposta",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_STRING),
		Help: &I18nObject{
			EnUS:   "Set a response format, ensure the output from llm is a valid code block as possible, such as JSON, XML, etc.",
			ZhHans: "设置一个返回格式，确保llm的输出尽可能是有效的代码块，如JSON、XML等",
			JaJp:   "応答形式を設定します。llmの出力が可能な限り有効なコードブロックであることを確認します。",
			PtBr:   "Defina um formato de resposta para garantir que a saída do llm seja um bloco de código válido o mais possível, como JSON, XML, etc.",
		},
		Required: false,
		Options:  []string{"JSON", "XML"},
	},
	JSON_SCHEMA: {
		Label: &I18nObject{
			EnUS: "JSON Schema",
		},
		Type: parser.ToPtr(PARAMETER_TYPE_STRING),
		Help: &I18nObject{
			EnUS:   "Set a response json schema will ensure LLM to adhere it.",
			ZhHans: "设置返回的json schema，llm将按照它返回",
		},
		Required: false,
	},
}

func (m *ModelParameterRule) TransformTemplate() error {
	if m.Label == nil || m.Label.EnUS == "" {
		m.Label = &I18nObject{
			EnUS: m.Name,
		}
	}

	// if use_template is not empty, transform to use default value
	if m.UseTemplate != nil && *m.UseTemplate != "" {
		// get the value of use_template
		useTemplateValue := m.UseTemplate
		// get the template
		template, ok := PARAMETER_RULE_TEMPLATE[DefaultParameterName(*useTemplateValue)]
		if !ok {
			return fmt.Errorf("use_template %s not found", *useTemplateValue)
		}
		// transform to default value
		if m.Label == nil {
			m.Label = template.Label
		}
		if m.Type == nil {
			m.Type = template.Type
		}
		if m.Help == nil {
			m.Help = template.Help
		}
		if m.Default == nil {
			m.Default = template.Default
		}
		if m.Min == nil {
			m.Min = template.Min
		}
		if m.Max == nil {
			m.Max = template.Max
		}
		if m.Precision == nil {
			m.Precision = template.Precision
		}
		if m.Options == nil {
			m.Options = template.Options
		}
	}
	if m.Options == nil {
		m.Options = []string{}
	}
	return nil
}

func (m *ModelParameterRule) UnmarshalJSON(data []byte) error {
	type alias ModelParameterRule

	temp := &struct {
		*alias
	}{
		alias: (*alias)(m),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if err := m.TransformTemplate(); err != nil {
		return err
	}

	return nil
}

func (m *ModelParameterRule) UnmarshalYAML(value *yaml.Node) error {
	type alias ModelParameterRule

	temp := &struct {
		*alias `yaml:",inline"`
	}{
		alias: (*alias)(m),
	}

	if err := value.Decode(&temp); err != nil {
		return err
	}

	if err := m.TransformTemplate(); err != nil {
		return err
	}

	return nil
}

func isParameterRule(fl validator.FieldLevel) bool {
	// if use_template is empty, then label, type should be required
	// try get the value of use_template
	useTemplateHandle := fl.Field().FieldByName("UseTemplate")
	// check if use_template is null pointer
	if useTemplateHandle.IsNil() {
		// label and type should be required
		// try get the value of label
		if fl.Field().FieldByName("Label").IsNil() {
			return false
		}

		// try get the value of type
		if fl.Field().FieldByName("Type").IsNil() {
			return false
		}
	}

	return true
}

type ModelPriceConfig struct {
	Input    decimal.Decimal  `json:"input" yaml:"input" validate:"required"`
	Output   *decimal.Decimal `json:"output" yaml:"output" validate:"omitempty"`
	Unit     decimal.Decimal  `json:"unit" yaml:"unit" validate:"required"`
	Currency string           `json:"currency" yaml:"currency" validate:"required"`
}

type ModelDeclaration struct {
	Model           string                         `json:"model" yaml:"model" validate:"required,lt=256"`
	Label           I18nObject                     `json:"label" yaml:"label" validate:"required"`
	ModelType       ModelType                      `json:"model_type" yaml:"model_type" validate:"required,model_type"`
	Features        []string                       `json:"features" yaml:"features" validate:"omitempty,lte=256,dive,lt=256"`
	FetchFrom       ModelProviderConfigurateMethod `json:"fetch_from" yaml:"fetch_from" validate:"omitempty,model_provider_configurate_method"`
	ModelProperties map[string]any                 `json:"model_properties" yaml:"model_properties" validate:"omitempty"`
	Deprecated      bool                           `json:"deprecated" yaml:"deprecated"`
	ParameterRules  []ModelParameterRule           `json:"parameter_rules" yaml:"parameter_rules" validate:"omitempty,lte=128,dive,parameter_rule"`
	PriceConfig     *ModelPriceConfig              `json:"pricing" yaml:"pricing" validate:"omitempty"`
}

func (m *ModelDeclaration) UnmarshalJSON(data []byte) error {
	type alias ModelDeclaration

	temp := &struct {
		*alias
	}{
		alias: (*alias)(m),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if m.FetchFrom == "" {
		m.FetchFrom = CONFIGURATE_METHOD_PREDEFINED_MODEL
	}

	if m.ParameterRules == nil {
		m.ParameterRules = []ModelParameterRule{}
	}

	return nil
}

func (m ModelDeclaration) MarshalJSON() ([]byte, error) {
	type alias ModelDeclaration

	temp := &struct {
		alias `json:",inline"`
	}{
		alias: (alias)(m),
	}

	if temp.Label.EnUS == "" {
		temp.Label.EnUS = temp.Model
	}

	// to avoid ModelProperties not serializable, we need to convert all the keys to string
	// includes inner map and slice
	if temp.ModelProperties != nil {
		result, ok := mapping.ConvertAnyMap(temp.ModelProperties).(map[string]any)
		if !ok {
			log.Error("ModelProperties is not a map[string]any", "model_properties", temp.ModelProperties)
		} else {
			temp.ModelProperties = result
		}
	}

	return json.Marshal(temp)
}

func (m *ModelDeclaration) UnmarshalYAML(value *yaml.Node) error {
	type alias ModelDeclaration

	temp := &struct {
		*alias `yaml:",inline"`
	}{
		alias: (*alias)(m),
	}

	if err := value.Decode(&temp); err != nil {
		return err
	}

	if m.FetchFrom == "" {
		m.FetchFrom = CONFIGURATE_METHOD_PREDEFINED_MODEL
	}

	if m.ParameterRules == nil {
		m.ParameterRules = []ModelParameterRule{}
	}

	return nil
}

type ModelProviderFormType string

const (
	FORM_TYPE_TEXT_INPUT   ModelProviderFormType = "text-input"
	FORM_TYPE_SECRET_INPUT ModelProviderFormType = "secret-input"
	FORM_TYPE_SELECT       ModelProviderFormType = "select"
	FORM_TYPE_RADIO        ModelProviderFormType = "radio"
	FORM_TYPE_SWITCH       ModelProviderFormType = "switch"
)

func isModelProviderFormType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(FORM_TYPE_TEXT_INPUT),
		string(FORM_TYPE_SECRET_INPUT),
		string(FORM_TYPE_SELECT),
		string(FORM_TYPE_RADIO),
		string(FORM_TYPE_SWITCH):
		return true
	}
	return false
}

type ModelProviderFormShowOnObject struct {
	Variable string `json:"variable" yaml:"variable" validate:"required,lt=256"`
	Value    string `json:"value" yaml:"value" validate:"required,lt=256"`
}

type ModelProviderFormOption struct {
	Label  I18nObject                      `json:"label" yaml:"label" validate:"required"`
	Value  string                          `json:"value" yaml:"value" validate:"required,lt=256"`
	ShowOn []ModelProviderFormShowOnObject `json:"show_on" yaml:"show_on" validate:"omitempty,lte=16,dive"`
}

func (m *ModelProviderFormOption) UnmarshalJSON(data []byte) error {
	// avoid show_on to be nil
	type Alias ModelProviderFormOption
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if m.ShowOn == nil {
		m.ShowOn = []ModelProviderFormShowOnObject{}
	}

	return nil
}

func (m *ModelProviderFormOption) UnmarshalYAML(value *yaml.Node) error {
	// avoid show_on to be nil
	type Alias ModelProviderFormOption
	aux := &struct {
		*Alias `yaml:",inline"`
	}{
		Alias: (*Alias)(m),
	}

	if err := value.Decode(&aux); err != nil {
		return err
	}

	if m.ShowOn == nil {
		m.ShowOn = []ModelProviderFormShowOnObject{}
	}

	return nil
}

type ModelProviderCredentialFormSchema struct {
	Variable    string                          `json:"variable" yaml:"variable" validate:"required,lt=256"`
	Label       I18nObject                      `json:"label" yaml:"label" validate:"required"`
	Type        ModelProviderFormType           `json:"type" yaml:"type" validate:"required,model_provider_form_type"`
	Required    bool                            `json:"required" yaml:"required"`
	Default     *string                         `json:"default" yaml:"default" validate:"omitempty,lt=256"`
	Options     []ModelProviderFormOption       `json:"options" yaml:"options" validate:"omitempty,lte=128,dive"`
	Placeholder *I18nObject                     `json:"placeholder" yaml:"placeholder" validate:"omitempty"`
	MaxLength   int                             `json:"max_length" yaml:"max_length"`
	ShowOn      []ModelProviderFormShowOnObject `json:"show_on" yaml:"show_on" validate:"omitempty,lte=16,dive"`
}

func (m *ModelProviderCredentialFormSchema) UnmarshalJSON(data []byte) error {
	type Alias ModelProviderCredentialFormSchema

	temp := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if m.ShowOn == nil {
		m.ShowOn = []ModelProviderFormShowOnObject{}
	}

	if m.Options == nil {
		m.Options = []ModelProviderFormOption{}
	}

	return nil
}

func (m *ModelProviderCredentialFormSchema) UnmarshalYAML(value *yaml.Node) error {
	type Alias ModelProviderCredentialFormSchema

	temp := &struct {
		*Alias `yaml:",inline"`
	}{
		Alias: (*Alias)(m),
	}

	if err := value.Decode(&temp); err != nil {
		return err
	}

	if m.ShowOn == nil {
		m.ShowOn = []ModelProviderFormShowOnObject{}
	}

	if m.Options == nil {
		m.Options = []ModelProviderFormOption{}
	}

	return nil
}

type ModelProviderCredentialSchema struct {
	CredentialFormSchemas []ModelProviderCredentialFormSchema `json:"credential_form_schemas" yaml:"credential_form_schemas" validate:"omitempty,lte=32,dive"`
}

type FieldModelSchema struct {
	Label       I18nObject  `json:"label" yaml:"label" validate:"required"`
	Placeholder *I18nObject `json:"placeholder" yaml:"placeholder" validate:"omitempty"`
}

type ModelCredentialSchema struct {
	Model                 FieldModelSchema                    `json:"model" yaml:"model" validate:"required"`
	CredentialFormSchemas []ModelProviderCredentialFormSchema `json:"credential_form_schemas" yaml:"credential_form_schemas" validate:"omitempty,lte=32,dive"`
}

type ModelProviderHelpEntity struct {
	Title I18nObject `json:"title" yaml:"title" validate:"required"`
	URL   I18nObject `json:"url" yaml:"url" validate:"required"`
}

type ModelPosition struct {
	LLM           *[]string `json:"llm,omitempty" yaml:"llm,omitempty"`
	TextEmbedding *[]string `json:"text_embedding,omitempty" yaml:"text_embedding,omitempty"`
	Rerank        *[]string `json:"rerank,omitempty" yaml:"rerank,omitempty"`
	TTS           *[]string `json:"tts,omitempty" yaml:"tts,omitempty"`
	Speech2text   *[]string `json:"speech2text,omitempty" yaml:"speech2text,omitempty"`
	Moderation    *[]string `json:"moderation,omitempty" yaml:"moderation,omitempty"`
}

type ModelProviderDeclaration struct {
	Provider                 string                           `json:"provider" yaml:"provider" validate:"required,lt=256"`
	Label                    I18nObject                       `json:"label" yaml:"label" validate:"required"`
	Description              *I18nObject                      `json:"description" yaml:"description,omitempty" validate:"omitempty"`
	IconSmall                *I18nObject                      `json:"icon_small" yaml:"icon_small,omitempty" validate:"omitempty"`
	IconLarge                *I18nObject                      `json:"icon_large" yaml:"icon_large,omitempty" validate:"omitempty"`
	Background               *string                          `json:"background" yaml:"background,omitempty" validate:"omitempty"`
	Help                     *ModelProviderHelpEntity         `json:"help" yaml:"help,omitempty" validate:"omitempty"`
	SupportedModelTypes      []ModelType                      `json:"supported_model_types" yaml:"supported_model_types" validate:"required,lte=16,dive,model_type"`
	ConfigurateMethods       []ModelProviderConfigurateMethod `json:"configurate_methods" yaml:"configurate_methods" validate:"required,lte=16,dive,model_provider_configurate_method"`
	ProviderCredentialSchema *ModelProviderCredentialSchema   `json:"provider_credential_schema" yaml:"provider_credential_schema,omitempty" validate:"omitempty"`
	ModelCredentialSchema    *ModelCredentialSchema           `json:"model_credential_schema" yaml:"model_credential_schema,omitempty" validate:"omitempty"`
	Position                 *ModelPosition                   `json:"position,omitempty" yaml:"position,omitempty"`
	Models                   []ModelDeclaration               `json:"models" yaml:"model_declarations,omitempty"`
	ModelFiles               []string                         `json:"-" yaml:"-"`
	PositionFiles            map[string]string                `json:"-" yaml:"-"`
}

func (m *ModelProviderDeclaration) UnmarshalJSON(data []byte) error {
	type alias ModelProviderDeclaration

	var temp struct {
		alias
		Models json.RawMessage `json:"models"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*m = ModelProviderDeclaration(temp.alias)

	if m.ModelCredentialSchema != nil && m.ModelCredentialSchema.CredentialFormSchemas == nil {
		m.ModelCredentialSchema.CredentialFormSchemas = []ModelProviderCredentialFormSchema{}
	}

	if m.ProviderCredentialSchema != nil && m.ProviderCredentialSchema.CredentialFormSchemas == nil {
		m.ProviderCredentialSchema.CredentialFormSchemas = []ModelProviderCredentialFormSchema{}
	}

	// unmarshal models into map[string]any
	var models map[string]any
	if err := json.Unmarshal(temp.Models, &models); err != nil {
		// can not unmarshal it into map, so it's a list
		if err := json.Unmarshal(temp.Models, &m.Models); err != nil {
			return err
		}

		return nil
	}

	m.PositionFiles = make(map[string]string)

	types := []string{
		"llm",
		"text_embedding",
		"tts",
		"speech2text",
		"moderation",
		"rerank",
	}

	for _, model_type := range types {
		modelTypeMap, ok := models[model_type].(map[string]any)
		if ok {
			modelTypePositionFile, ok := modelTypeMap["position"]
			if ok {
				modelTypePositionFilePath, ok := modelTypePositionFile.(string)
				if ok {
					m.PositionFiles[model_type] = modelTypePositionFilePath
				}
			}

			modelTypePredefinedFiles, ok := modelTypeMap["predefined"].([]string)
			if ok {
				m.ModelFiles = append(m.ModelFiles, modelTypePredefinedFiles...)
			}
		}
	}

	if m.Models == nil {
		m.Models = []ModelDeclaration{}
	}

	return nil
}

func (m *ModelProviderDeclaration) MarshalJSON() ([]byte, error) {
	type alias ModelProviderDeclaration

	temp := &struct {
		*alias `json:",inline"`
	}{
		alias: (*alias)(m),
	}

	if temp.Models == nil {
		temp.Models = []ModelDeclaration{}
	}

	return json.Marshal(temp)
}

func (m *ModelProviderDeclaration) UnmarshalYAML(value *yaml.Node) error {
	type alias ModelProviderDeclaration

	var temp struct {
		alias  `yaml:",inline"`
		Models yaml.Node `yaml:"models"`
	}

	if err := value.Decode(&temp); err != nil {
		return err
	}

	*m = ModelProviderDeclaration(temp.alias)

	if m.ModelCredentialSchema != nil && m.ModelCredentialSchema.CredentialFormSchemas == nil {
		m.ModelCredentialSchema.CredentialFormSchemas = []ModelProviderCredentialFormSchema{}
	}

	if m.ProviderCredentialSchema != nil && m.ProviderCredentialSchema.CredentialFormSchemas == nil {
		m.ProviderCredentialSchema.CredentialFormSchemas = []ModelProviderCredentialFormSchema{}
	}

	// Check if Models is a mapping node
	if temp.Models.Kind == yaml.MappingNode {
		m.PositionFiles = make(map[string]string)

		types := []string{
			"llm",
			"text_embedding",
			"tts",
			"speech2text",
			"moderation",
			"rerank",
		}

		for i := 0; i < len(temp.Models.Content); i += 2 {
			key := temp.Models.Content[i].Value
			value := temp.Models.Content[i+1]

			for _, model_type := range types {
				if key == model_type {
					if value.Kind == yaml.MappingNode {
						for j := 0; j < len(value.Content); j += 2 {
							if value.Content[j].Value == "position" {
								m.PositionFiles[model_type] = value.Content[j+1].Value
							} else if value.Content[j].Value == "predefined" {
								// get content of predefined
								if value.Content[j+1].Kind == yaml.SequenceNode {
									for _, file := range value.Content[j+1].Content {
										m.ModelFiles = append(m.ModelFiles, file.Value)
									}
								}
							}
						}
					}
				}
			}
		}
	} else if temp.Models.Kind == yaml.SequenceNode {
		if err := temp.Models.Decode(&m.Models); err != nil {
			return err
		}
	}

	if m.Models == nil {
		m.Models = []ModelDeclaration{}
	}

	return nil
}

func init() {
	// init validator
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")
	// register translations for default validators
	en_translations.RegisterDefaultTranslations(validators.GlobalEntitiesValidator, translator)

	validators.GlobalEntitiesValidator.RegisterValidation("model_type", isModelType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"model_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_type", "{0} is not a valid model type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_type", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("model_provider_configurate_method", isModelProviderConfigurateMethod)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"model_provider_configurate_method",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_provider_configurate_method", "{0} is not a valid model provider configurate method", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_provider_configurate_method", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("model_provider_form_type", isModelProviderFormType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"model_provider_form_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_provider_form_type", "{0} is not a valid model provider form type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_provider_form_type", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("model_parameter_type", isModelParameterType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"model_parameter_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_parameter_type", "{0} is not a valid model parameter type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_parameter_type", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("parameter_rule", isParameterRule)

	validators.GlobalEntitiesValidator.RegisterValidation("is_basic_type", isBasicType)
}
