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

type OnlineDocumentPageChunk struct {
	Result []map[string]any `json:"result"`
}

// Online driver file structures
type OnlineDriverFilePath struct {
	Key  string `json:"key" validate:"required"`  // The key of the file
	Size int    `json:"size" validate:"required"` // The size of the file
}

type OnlineDriverFileBucket struct {
	Bucket      string                 `json:"bucket" validate:"required"`       // The bucket of the file
	Files       []OnlineDriverFilePath `json:"files" validate:"required"`        // The files of the bucket
	IsTruncated bool                   `json:"is_truncated" validate:"required"` // Whether the bucket has more files
}

type GetOnlineDriverFileListResponse struct {
	Result []OnlineDriverFileBucket `json:"result" validate:"required"` // The bucket of the files
}

type OnlineDriverFile struct {
	ID        string `json:"id" validate:"required"`         // The id of the file
	TenantID  string `json:"tenant_id" validate:"required"`  // The tenant id of the file
	Type      string `json:"type" validate:"required"`       // The type of the file
	RemoteURL string `json:"remote_url" validate:"required"` // The remote url of the file
	RelatedID string `json:"related_id" validate:"required"` // The related id of the file
	Filename  string `json:"filename" validate:"required"`   // The name of the file
	Extension string `json:"extension" validate:"required"`  // The extension of the file
	MimeType  string `json:"mime_type" validate:"required"`  // The mime type of the file
	Size      int    `json:"size" validate:"required"`       // The size of the file
}
