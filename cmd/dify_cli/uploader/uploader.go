package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/encryption"
)

func signRequest(secret string, timestamp string, body []byte) string {
	data := append([]byte(timestamp+"."), body...)
	return "sha256=" + encryption.HmacSha256(secret, data)
}

func getSignedURL(cfg *types.DifyConfig, filename, mimetype string) (string, error) {
	reqBody := map[string]string{
		"filename": filename,
		"mimetype": mimetype,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := signRequest(cfg.Env.CliApiSecret, timestamp, body)

	url := strings.TrimSuffix(cfg.Env.CliApiURL, "/") + "/cli/api/upload/file/request"

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cli-Api-Session-Id", cfg.Env.CliApiSessionID)
	req.Header.Set("X-Cli-Api-Timestamp", timestamp)
	req.Header.Set("X-Cli-Api-Signature", signature)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get signed URL: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result types.DifyInnerAPIResponse[types.SignedURLResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("API error: %s", result.Error)
	}

	if result.Data == nil {
		return "", fmt.Errorf("no data in response")
	}

	return result.Data.URL, nil
}

func uploadFileToSignedURL(signedURL string, filePath string, filename string) (*types.FileUploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}

	req, err := http.NewRequest("POST", signedURL, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result types.FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &result, nil
}

func UploadFile(cfg *types.DifyConfig, filePath string) (*types.ToolFileObject, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	filename := filepath.Base(absPath)
	ext := filepath.Ext(filename)
	mimetype := mime.TypeByExtension(ext)
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}

	signedURL, err := getSignedURL(cfg, filename, mimetype)
	if err != nil {
		return nil, fmt.Errorf("failed to get signed URL: %w", err)
	}

	uploadResp, err := uploadFileToSignedURL(signedURL, absPath, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	fileType := detectFileType(mimetype)

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &types.ToolFileObject{
		DifyModelIdentity: types.DifyFileIdentity,
		URL:               uploadResp.PreviewURL,
		MimeType:          mimetype,
		Filename:          filename,
		Extension:         ext,
		Size:              int(fileInfo.Size()),
		Type:              fileType,
	}, nil
}

func detectFileType(mimetype string) types.FileType {
	if strings.HasPrefix(mimetype, "image/") {
		return types.FileTypeImage
	}
	if strings.HasPrefix(mimetype, "audio/") {
		return types.FileTypeAudio
	}
	if strings.HasPrefix(mimetype, "video/") {
		return types.FileTypeVideo
	}

	documentTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument",
		"text/",
		"application/json",
		"application/xml",
	}
	for _, docType := range documentTypes {
		if strings.HasPrefix(mimetype, docType) || strings.Contains(mimetype, docType) {
			return types.FileTypeDocument
		}
	}

	return types.FileTypeCustom
}
