package datasource_entities

type DataSourceValidateCredentialsResponse struct {
	Result bool `json:"result"`
}

type DataSourceInvokeFirstStepResponse struct {
	Result []map[string]any `json:"result"`
}

type DataSourceInvokeSecondStepResponse struct {
	Result []map[string]any `json:"result"`
}

type DatasourceInvokeOnlineDocumentGetContentResponse struct {
	Result map[string]any `json:"result"`
}
