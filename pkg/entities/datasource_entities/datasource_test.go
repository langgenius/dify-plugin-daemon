package datasource_entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnlineDriveFileRemoteMetadataIsOptional(t *testing.T) {
	file := OnlineDriveFile{
		ID:   "file-1",
		Name: "report.pdf",
		Size: 42,
		Type: "file",
	}

	encoded, err := json.Marshal(file)
	require.NoError(t, err)
	require.JSONEq(t, `{"id":"file-1","name":"report.pdf","size":42,"type":"file"}`, string(encoded))
}

func TestOnlineDriveBrowseResponsePreservesRemoteMetadata(t *testing.T) {
	payload := []byte(`
{
  "result": [
    {
      "bucket": "bucket-1",
      "files": [
        {
          "id": "file-1",
          "name": "report.pdf",
          "size": 42,
          "type": "file",
          "remote_metadata": {
            "version_id": "version-1",
            "etag": "etag-1",
            "checksum": {"algorithm": "QuickXorHash", "value": "checksum-1"},
            "modified_time": "2026-09-17T10:20:30Z"
          }
        }
      ],
      "is_truncated": false,
      "next_page_parameters": null
    }
  ]
}`)

	var response GetOnlineDriveBrowseFilesResponse
	require.NoError(t, json.Unmarshal(payload, &response))
	require.Len(t, response.Result, 1)
	require.Len(t, response.Result[0].Files, 1)

	metadata := response.Result[0].Files[0].RemoteMetadata
	require.NotNil(t, metadata)
	require.NotNil(t, metadata.VersionID)
	require.NotNil(t, metadata.ETag)
	require.NotNil(t, metadata.Checksum)
	require.NotNil(t, metadata.ModifiedTime)
	require.Equal(t, "version-1", *metadata.VersionID)
	require.Equal(t, "etag-1", *metadata.ETag)
	require.Equal(t, "QuickXorHash", metadata.Checksum.Algorithm)
	require.Equal(t, "checksum-1", metadata.Checksum.Value)
	require.Equal(t, "2026-09-17T10:20:30Z", *metadata.ModifiedTime)

	encoded, err := json.Marshal(response)
	require.NoError(t, err)
	require.JSONEq(t, string(payload), string(encoded))
}

func TestDataSourceResponseChunkPreservesRemoteMetadata(t *testing.T) {
	payload := []byte(`
{
  "type": "blob",
  "message": {"blob": "Y29udGVudA=="},
  "meta": {
    "remote_metadata": {
      "etag": "etag-1",
      "checksum": {"algorithm": "CRC32C", "value": "checksum-1"}
    }
  }
}`)

	var chunk DataSourceResponseChunk
	require.NoError(t, json.Unmarshal(payload, &chunk))

	encoded, err := json.Marshal(chunk)
	require.NoError(t, err)
	require.JSONEq(t, string(payload), string(encoded))
}
