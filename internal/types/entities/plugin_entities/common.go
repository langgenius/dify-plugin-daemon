package plugin_entities

type I18nObject struct {
	EnUS   string `json:"en_US" validate:"required,gt=0,lt=1024"`
	ZhHans string `json:"zh_Hans" validate:"lt=1024"`
	PtBr   string `json:"pt_BR" validate:"lt=1024"`
}
