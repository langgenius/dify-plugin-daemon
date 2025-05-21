package requests

type BaseRequestInvokeDatasource struct {
	Provider   string `json:"provider" validate:"required"`
	Datasource string `json:"datasource" validate:"required"`
}

type RequestValidateDatasourceCredentials struct {
	Credentials

	Provider string `json:"provider" validate:"required"`
}

type RequestInvokeDatasourceFirstStep struct {
	Credentials
	BaseRequestInvokeDatasource

	DatasourceParameters map[string]any `json:"datasource_parameters" validate:"required"`
}

type RequestInvokeDatasourceSecondStep struct {
	Credentials
	BaseRequestInvokeDatasource

	DatasourceParameters map[string]any `json:"datasource_parameters" validate:"required"`
}
