package datasource_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type DataSourceValidateCredentialsResponse struct {
	Result bool `json:"result"`
}

type DataSourceResponseChunkType string

const (
	DataSourceResponseChunkTypeText               DataSourceResponseChunkType = "text"
	DataSourceResponseChunkTypeFile               DataSourceResponseChunkType = "file"
	DataSourceResponseChunkTypeBlob               DataSourceResponseChunkType = "blob"
	DataSourceResponseChunkTypeBlobChunk          DataSourceResponseChunkType = "blob_chunk"
	DataSourceResponseChunkTypeJson               DataSourceResponseChunkType = "json"
	DataSourceResponseChunkTypeLink               DataSourceResponseChunkType = "link"
	DataSourceResponseChunkTypeImage              DataSourceResponseChunkType = "image"
	DataSourceResponseChunkTypeImageLink          DataSourceResponseChunkType = "image_link"
	DataSourceResponseChunkTypeVariable           DataSourceResponseChunkType = "variable"
	DataSourceResponseChunkTypeLog                DataSourceResponseChunkType = "log"
	DataSourceResponseChunkTypeRetrieverResources DataSourceResponseChunkType = "retriever_resources"
)

func IsValidDataSourceResponseChunkType(fl validator.FieldLevel) bool {
	t := fl.Field().String()
	switch DataSourceResponseChunkType(t) {
	case DataSourceResponseChunkTypeText,
		DataSourceResponseChunkTypeFile,
		DataSourceResponseChunkTypeBlob,
		DataSourceResponseChunkTypeBlobChunk,
		DataSourceResponseChunkTypeJson,
		DataSourceResponseChunkTypeLink,
		DataSourceResponseChunkTypeImage,
		DataSourceResponseChunkTypeImageLink,
		DataSourceResponseChunkTypeVariable,
		DataSourceResponseChunkTypeLog,
		DataSourceResponseChunkTypeRetrieverResources:
		return true
	default:
		return false
	}
}

func init() {
	err := validators.GlobalEntitiesValidator.RegisterValidation(
		"is_valid_data_source_response_chunk_type",
		IsValidDataSourceResponseChunkType,
	)
	if err != nil {
		panic(err)
	}
}

type DataSourceResponseChunk struct {
	Type    DataSourceResponseChunkType `json:"type" validate:"required,is_valid_data_source_response_chunk_type"`
	Message map[string]any              `json:"message"`
	Meta    map[string]any              `json:"meta"`
}

type WebsiteCrawlChunk struct {
	Result map[string]any `json:"result"`
}

type DatasourceGetPagesResponse struct {
	Result []map[string]any `json:"result"`
}

type OnlineDriveFile struct {
	ID   string `json:"id" validate:"required"`   // The file ID
	Name string `json:"name" validate:"required"` // The file name
	Size int    `json:"size" validate:"required"` // The file size
	Type string `json:"type" validate:"required"` // The file type: folder or file
}

type OnlineDriveFileBucket struct {
	Bucket             *string                `json:"bucket" validate:"omitempty"`               // The file bucket (optional)
	Files              []OnlineDriveFile      `json:"files" validate:"required"`                 // The file list
	IsTruncated        bool                   `json:"is_truncated" validate:"required"`          // Whether the result is truncated
	NextPageParameters map[string]interface{} `json:"next_page_parameters" validate:"omitempty"` // Parameters for fetching the next page
}

type GetOnlineDriveBrowseFilesResponse struct {
	Result []OnlineDriveFileBucket `json:"result" validate:"required"` // The list of file buckets
}
