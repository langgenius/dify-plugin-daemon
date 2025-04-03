package azure

import (
	"bytes"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type AzureBlobStorage struct {
	client        *azblob.Client
	containerName string
}

func NewAzureBlobStorage(containerName string, connectionString string) (oss.OSS, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}

	return &AzureBlobStorage{
		client:        client,
		containerName: containerName,
	}, nil
}

func (a *AzureBlobStorage) Save(key string, data []byte) error {
	_, err := a.client.UploadBuffer(context.TODO(), a.containerName, key, data, nil)
	return err
}

func (a *AzureBlobStorage) Load(key string) ([]byte, error) {
	get, err := a.client.DownloadStream(context.TODO(), a.containerName, key, nil)
	if err != nil {
		return nil, err
	}

	downloadedData := bytes.Buffer{}
	retryReader := get.NewRetryReader(context.TODO(), &azblob.RetryReaderOptions{})
	_, err = downloadedData.ReadFrom(retryReader)
	if err != nil {
		return nil, err
	}

	err = retryReader.Close()
	if err != nil {
		return nil, err
	}

	return downloadedData.Bytes(), nil
}

func (a *AzureBlobStorage) Exists(key string) (bool, error) {
	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(key)
	_, err := blobClient.GetProperties(context.TODO(), nil)

	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (a *AzureBlobStorage) State(key string) (oss.OSSState, error) {
	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(key)
	props, err := blobClient.GetProperties(context.TODO(), nil)

	if err != nil {
		return oss.OSSState{}, err
	}

	return oss.OSSState{
		Size:         *props.ContentLength,
		LastModified: *props.LastModified,
	}, nil
}

func (a *AzureBlobStorage) List(prefix string) ([]oss.OSSPath, error) {
	// append a slash to the prefix if it doesn't end with one
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	pager := a.client.NewListBlobsFlatPager(a.containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	paths := make([]oss.OSSPath, 0)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, blob := range page.Segment.BlobItems {
			// remove prefix
			key := strings.TrimPrefix(*blob.Name, prefix)
			// remove leading slash
			key = strings.TrimPrefix(key, "/")
			paths = append(paths, oss.OSSPath{
				Path:  key,
				IsDir: false,
			})
		}
	}

	return paths, nil
}

func (a *AzureBlobStorage) Delete(key string) error {
	_, err := a.client.DeleteBlob(context.TODO(), a.containerName, key, nil)
	return err
}

func (a *AzureBlobStorage) Type() string {
	return oss.OSS_TYPE_AZURE_BLOB
}
