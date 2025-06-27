package requests

type BaseRequestInvokeDatasource struct {
	Provider   string `json:"provider" validate:"required"`
	Datasource string `json:"datasource" validate:"required"`
}

type RequestValidateDatasourceCredentials struct {
	Credentials

	Provider string `json:"provider" validate:"required"`
}

type RequestInvokeDatasourceRequest struct {
	Credentials
	BaseRequestInvokeDatasource

	DatasourceParameters map[string]any `json:"datasource_parameters" validate:"required"`
}

type RequestDatasourceGetWebsiteCrawl RequestInvokeDatasourceRequest
type RequestDatasourceGetOnlineDocumentPages RequestInvokeDatasourceRequest

type RequestInvokeOnlineDocumentDatasourceGetContent struct {
	Credentials
	BaseRequestInvokeDatasource

	Page map[string]any `json:"page" validate:"required"`
}

// Online driver file request structures
type RequestGetOnlineDriverFileList struct {
	Credentials
	BaseRequestInvokeDatasource

	Prefix     *string `json:"prefix" validate:"omitempty"`      // File path prefix for filtering eg: 'docs/dify/'
	Bucket     *string `json:"bucket" validate:"omitempty"`      // Storage bucket name
	MaxKeys    int     `json:"max_keys" validate:"required"`     // Maximum number of files to return
	StartAfter *string `json:"start_after" validate:"omitempty"` // Pagination token for continuing from a specific file eg: 'docs/dify/1.txt'
}

type RequestGetOnlineDriverFile struct {
	Credentials
	BaseRequestInvokeDatasource

	Key    string  `json:"key" validate:"required"`     // The name of the file
	Bucket *string `json:"bucket" validate:"omitempty"` // The name of the bucket
}
