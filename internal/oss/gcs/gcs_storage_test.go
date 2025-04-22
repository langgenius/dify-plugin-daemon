package gcs_test

import (
	"context"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/oss/gcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func getRandomBucketName(t *testing.T) string {
	t.Helper()
	bucketName := "test-bucket-" + uuid.NewString()
	return bucketName
}

func setupTestGCS(t *testing.T, bucketName string, initialObjects []fakestorage.Object) *gcs.GCSStorage {
	t.Helper()
	ctx := context.Background()
	// Create the bucket
	fakeServer.CreateBucketWithOpts(
		fakestorage.CreateBucketOpts{
			Name: bucketName,
		},
	)
	// Create initial objects if provided
	for _, obj := range initialObjects {
		require.Equal(t, obj.ObjectAttrs.BucketName, bucketName, "Object must belong to the created bucket")
		fakeServer.CreateObject(obj)
	}

	// Create the GCSStorage instance using the fake server's endpoint
	storageInstance, err := gcs.NewGCSStorage(ctx, bucketName, option.WithHTTPClient(fakeServer.HTTPClient()), option.WithCredentials(&google.Credentials{}))
	require.NoError(t, err, "Failed to create GCSStorage instance")

	return storageInstance
}

func TestGCSStorage_Type(t *testing.T) {
	bucketName := getRandomBucketName(t)
	storageInstance := setupTestGCS(t, bucketName, []fakestorage.Object{})
	assert.Equal(t, oss.OSS_TYPE_GCS, storageInstance.Type())
}

func TestGCSStorage_Load(t *testing.T) {
	bucketName := getRandomBucketName(t)
	storageInstance := setupTestGCS(t, bucketName, []fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "file1.txt",
				BucketName: bucketName,
			},
			Content: []byte("file1"),
		},
	})

	actual, err := storageInstance.Load("file1.txt")
	require.NoError(t, err, "Load should succeed for existing file")
	assert.Equal(t, "file1", string(actual))
}

func TestGCSStorage_Exists(t *testing.T) {
	bucketName := getRandomBucketName(t)
	storageInstance := setupTestGCS(t, bucketName, []fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "file1.txt",
				BucketName: bucketName,
			},
			Content: []byte("file1"),
		},
	})

	tests := map[string]struct {
		key      string
		expected bool
	}{
		"FileExists": {
			key:      "file1.txt",
			expected: true,
		},
		"FileDoesNotExist": {
			key:      "non_existent_file.txt",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exists, err := storageInstance.Exists(tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, exists)
		})
	}
}

func TestGCSStorage_State(t *testing.T) {
	bucketName := getRandomBucketName(t)
	storageInstance := setupTestGCS(t, bucketName, []fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "file1.txt",
				BucketName: bucketName,
			},
			Content: []byte("file1"),
		},
	})

	state, err := storageInstance.State("file1.txt")
	require.NoError(t, err, "State should succeed for existing file")
	assert.Greater(t, state.Size, int64(0), "File size should be greater than 0")
	assert.NotZero(t, state.LastModified, "Last modified time should not be zero")
}

func TestGCSStorage_List(t *testing.T) {
	bucketName := getRandomBucketName(t)
	storageInstance := setupTestGCS(t, bucketName, []fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "file1.txt",
				BucketName: bucketName,
			},
			Content: []byte("file1"),
		},
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "dir1/file2.txt",
				BucketName: bucketName,
			},
			Content: []byte("file2"),
		},
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "dir1/subdir/file3.txt",
				BucketName: bucketName,
			},
			Content: []byte("file3"),
		},
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "dir2/file4.txt",
				BucketName: bucketName,
			},
			Content: []byte("file4"),
		},
	})

	tests := map[string]struct {
		prefix   string
		expected []oss.OSSPath
	}{
		"ListRootDirectoryWithoutSlask": {
			prefix: "",
			expected: []oss.OSSPath{
				{Path: "file1.txt", IsDir: false},
				{Path: "dir1/file2.txt", IsDir: false},
				{Path: "dir1/subdir/file3.txt", IsDir: false},
				{Path: "dir2/file4.txt", IsDir: false},
			},
		},
		"ListRootDirectoryWithSlash": {
			prefix: "/",
			expected: []oss.OSSPath{
				{Path: "file1.txt", IsDir: false},
				{Path: "dir1/file2.txt", IsDir: false},
				{Path: "dir1/subdir/file3.txt", IsDir: false},
				{Path: "dir2/file4.txt", IsDir: false},
			},
		},
		"ListDirectoryWithSlash": {
			prefix: "dir1/",
			expected: []oss.OSSPath{
				{Path: "file2.txt", IsDir: false},
				{Path: "subdir/file3.txt", IsDir: false},
			},
		},
		"ListDirectoryWithoutSlash": {
			prefix: "dir1",
			expected: []oss.OSSPath{
				{Path: "file2.txt", IsDir: false},
				{Path: "subdir/file3.txt", IsDir: false},
			},
		},
		"ListNonExistentDirectory": {
			prefix:   "non_existent_dir/",
			expected: []oss.OSSPath{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actualRoot, err := storageInstance.List(tt.prefix)
			require.NoError(t, err, "List with prefix should succeed")
			assert.ElementsMatch(t, tt.expected, actualRoot, "List(\"%s\") should return expected files and dirs", tt.prefix)
		})
	}
}

func TestGCSStorage_Delete(t *testing.T) {
	bucketName := getRandomBucketName(t)
	storageInstance := setupTestGCS(t, bucketName, []fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				Name:       "file_to_delete.txt",
				BucketName: bucketName,
			},
			Content: []byte("file to be deleted"),
		},
	})

	err := storageInstance.Delete("file_to_delete.txt")
	require.NoError(t, err, "Delete should succeed for existing file")
	// Verify file doesn't exist after deletion
	exists, err := storageInstance.Exists("file_to_delete.txt")
	require.NoError(t, err)
	assert.False(t, exists, "File should not exist after deletion")
}
