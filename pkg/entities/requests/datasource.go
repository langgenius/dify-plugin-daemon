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

type DatasourceOnlineDriveBrowseFilesRequest struct {
	Credentials
	BaseRequestInvokeDatasource

	Request OnlineDriveBrowseFilesRequest `json:"request" validate:"required"`
}

type OnlineDriveBrowseFilesRequest struct {
	Bucket             *string                `json:"bucket" validate:"omitempty"`               // The file bucket (optional)
	Prefix             string                 `json:"prefix" validate:"omitempty"`               // The parent folder ID
	MaxKeys            int                    `json:"max_keys" validate:"required"`              // Page size for pagination
	NextPageParameters map[string]interface{} `json:"next_page_parameters" validate:"omitempty"` // Parameters for fetching the next page
}

type OnlineDriveDownloadFileRequest struct {
	Bucket *string `json:"bucket" validate:"omitempty"` // The file bucket (optional)
	ID     string  `json:"id" validate:"required"`      // The file ID
}

type DatasourceOnlineDriveDownloadFileRequest struct {
	Credentials
	BaseRequestInvokeDatasource

	Request OnlineDriveDownloadFileRequest `json:"request" validate:"required"`
}
