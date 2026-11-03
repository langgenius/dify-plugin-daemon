package controllers

import (
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	daemonlog "github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/stretchr/testify/require"
)

func TestUploadHandlersRemoveMultipartTempFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	invalidPayload := bytes.Repeat([]byte("not-a-zip"), 64)
	var bundlePayload bytes.Buffer
	bundleWriter := zip.NewWriter(&bundlePayload)
	manifest, err := bundleWriter.Create("manifest.yaml")
	require.NoError(t, err)
	_, err = manifest.Write([]byte("version: 0.0.1\nname: test-bundle\ndependencies: []\n"))
	require.NoError(t, err)
	require.NoError(t, bundleWriter.Close())

	tests := []struct {
		name              string
		bundle            bool
		tenantID          string
		multipart         bool
		formField         string
		maxPackageSize    int64
		payload           []byte
		wantResponseMatch string
	}{
		{
			name:              "plugin rejected above size limit",
			tenantID:          "tenant-1",
			multipart:         true,
			formField:         "dify_pkg",
			maxPackageSize:    128,
			wantResponseMatch: "file size exceeds the maximum limit",
		},
		{
			name:              "bundle rejected above size limit",
			bundle:            true,
			tenantID:          "tenant-1",
			multipart:         true,
			formField:         "dify_bundle",
			maxPackageSize:    128,
			wantResponseMatch: "file size exceeds the maximum limit",
		},
		{
			name:              "expected file field missing",
			tenantID:          "tenant-1",
			multipart:         true,
			formField:         "other_file",
			maxPackageSize:    int64(len(invalidPayload) + 1),
			wantResponseMatch: "no such file",
		},
		{
			name:              "tenant ID missing",
			multipart:         true,
			formField:         "dify_pkg",
			maxPackageSize:    int64(len(invalidPayload) + 1),
			wantResponseMatch: "tenant ID is required",
		},
		{
			name:              "request is not multipart",
			tenantID:          "tenant-1",
			maxPackageSize:    int64(len(invalidPayload) + 1),
			wantResponseMatch: "request Content-Type isn't multipart/form-data",
		},
		{
			name:              "plugin archive is invalid",
			tenantID:          "tenant-1",
			multipart:         true,
			formField:         "dify_pkg",
			maxPackageSize:    int64(len(invalidPayload) + 1),
			wantResponseMatch: "not a valid difypkg file",
		},
		{
			name:              "bundle accepted",
			bundle:            true,
			tenantID:          "tenant-1",
			multipart:         true,
			formField:         "dify_bundle",
			maxPackageSize:    int64(bundlePayload.Len() + 1),
			payload:           bundlePayload.Bytes(),
			wantResponseMatch: `"code":0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			must := require.New(t)
			payload := tt.payload
			if payload == nil {
				payload = invalidPayload
			}
			tempDir := t.TempDir()
			t.Setenv("TMPDIR", tempDir)

			config := &app.Config{
				MaxPluginPackageSize: tt.maxPackageSize,
				MaxBundlePackageSize: tt.maxPackageSize,
			}
			engine := gin.New()
			engine.MaxMultipartMemory = 64
			engine.Use(daemonlog.TraceMiddleware())

			type formObservation struct {
				parseErr  error
				fileCount int
				formNil   bool
			}
			observed := make(chan formObservation, 1)
			handled := make(chan struct{}, 1)
			engine.Use(func(c *gin.Context) {
				observation := formObservation{formNil: c.Request.MultipartForm == nil}
				if tt.multipart {
					observation.parseErr = c.Request.ParseMultipartForm(engine.MaxMultipartMemory)
					files, err := filepath.Glob(filepath.Join(tempDir, "multipart-*"))
					if err != nil && observation.parseErr == nil {
						observation.parseErr = err
					}
					observation.fileCount = len(files)
				}
				observed <- observation
				c.Next()
				handled <- struct{}{}
			})

			route := "/upload"
			requestPath := route
			if tt.tenantID != "" {
				route = "/upload/:tenant_id"
				requestPath = "/upload/" + tt.tenantID
			}
			if tt.bundle {
				engine.POST(route, UploadBundle(config))
			} else {
				engine.POST(route, UploadPlugin(config))
			}

			server := httptest.NewServer(engine)
			t.Cleanup(server.Close)

			var body bytes.Buffer
			contentType := ""
			if tt.multipart {
				writer := multipart.NewWriter(&body)
				part, err := writer.CreateFormFile(tt.formField, "upload.difypkg")
				must.NoError(err)
				_, err = part.Write(payload)
				must.NoError(err)
				must.NoError(writer.Close())
				contentType = writer.FormDataContentType()
			}

			request, err := http.NewRequest(http.MethodPost, server.URL+requestPath, &body)
			must.NoError(err)
			if contentType != "" {
				request.Header.Set("Content-Type", contentType)
			}
			response, err := http.DefaultClient.Do(request)
			must.NoError(err)
			responseBody, err := io.ReadAll(response.Body)
			must.NoError(err)
			must.NoError(response.Body.Close())
			<-handled

			must.Equal(http.StatusOK, response.StatusCode)
			must.Contains(string(responseBody), tt.wantResponseMatch)
			observation := <-observed
			must.NoError(observation.parseErr)
			must.True(observation.formNil)
			if tt.multipart {
				must.Equal(1, observation.fileCount, "test payload must spill to a multipart temp file")
			} else {
				must.Zero(observation.fileCount)
			}

			files, err := filepath.Glob(filepath.Join(tempDir, "multipart-*"))
			must.NoError(err)
			must.Empty(files, "multipart temp files remain after %s", tt.name)
		})
	}
}

func TestListPluginsByCategoryRequestBindsFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var received listPluginsByCategoryRequest
	router.GET("/plugin/:tenant_id/:category/list", func(c *gin.Context) {
		BindRequest(c, func(request listPluginsByCategoryRequest) {
			received = request
			c.Status(http.StatusNoContent)
		})
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/plugin/tenant-1/tool/list?page=2&page_size=25&query=Weather&tags=search&tags=rag&language=zh_Hans",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "tenant-1", received.TenantID)
	require.Equal(t, plugin_entities.PLUGIN_CATEGORY_TOOL, received.Category)
	require.Equal(t, 2, received.Page)
	require.Equal(t, 25, received.PageSize)
	require.Equal(t, "Weather", received.Query)
	require.Equal(
		t,
		[]manifest_entities.PluginTag{
			manifest_entities.PLUGIN_TAG_SEARCH,
			manifest_entities.PLUGIN_TAG_RAG,
		},
		received.Tags,
	)
	require.Equal(t, "zh_Hans", received.Language)
}

func TestListInstalledPluginIDsRequestBindsCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var received listInstalledPluginIDsRequest
	router.GET("/plugin/:tenant_id/installation/ids", func(c *gin.Context) {
		BindRequest(c, func(request listInstalledPluginIDsRequest) {
			received = request
			c.Status(http.StatusNoContent)
		})
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/plugin/tenant-1/installation/ids?category=tool",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "tenant-1", received.TenantID)
	require.Equal(t, plugin_entities.PLUGIN_CATEGORY_TOOL, received.Category)
}
