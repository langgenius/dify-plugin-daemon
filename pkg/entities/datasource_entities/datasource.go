package datasource_entities

type DataSourceValidateCredentialsResponse struct {
	Result bool `json:"result"`
}

type DataSourceInvokeFirstStepResponse struct {
	Response map[string]any `json:"response"`
}

type DataSourceInvokeSecondStepResponse struct {
	Response map[string]any `json:"response"`
}
