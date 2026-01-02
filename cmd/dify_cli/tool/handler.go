package tool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
)

type HandlerFunc func(chunk *types.DifyToolResponseChunk, w io.Writer) error

type Registry struct {
	handlers map[types.DifyToolResponseChunkType]HandlerFunc
}

func NewRegistry() *Registry {
	r := &Registry{handlers: make(map[types.DifyToolResponseChunkType]HandlerFunc)}
	r.Register(types.ToolResponseChunkTypeText, handleText)
	r.Register(types.ToolResponseChunkTypeJson, handleJSON)
	r.Register(types.ToolResponseChunkTypeLink, handleLink)
	r.Register(types.ToolResponseChunkTypeImage, handleImage)
	r.Register(types.ToolResponseChunkTypeImageLink, handleImageLink)
	r.Register(types.ToolResponseChunkTypeFile, handleFile)
	r.Register(types.ToolResponseChunkTypeBlob, handleBlob)
	r.Register(types.ToolResponseChunkTypeBlobChunk, handleBlobChunk)
	r.Register(types.ToolResponseChunkTypeBinaryLink, handleBinaryLink)
	r.Register(types.ToolResponseChunkTypeVariable, handleVariable)
	r.Register(types.ToolResponseChunkTypeLog, handleLog)
	r.Register(types.ToolResponseChunkTypeRetrieverResources, handleRetrieverResources)
	return r
}

func (r *Registry) Register(t types.DifyToolResponseChunkType, h HandlerFunc) {
	r.handlers[t] = h
}

func (r *Registry) Dispatch(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	if h, ok := r.handlers[chunk.Type]; ok {
		return h(chunk, w)
	}
	data, _ := json.Marshal(chunk)
	fmt.Fprintf(w, "[%s] %s\n", chunk.Type, string(data))
	return nil
}

var globalRegistry = NewRegistry()

func Dispatch(chunk *types.DifyToolResponseChunk) error {
	return globalRegistry.Dispatch(chunk, os.Stdout)
}

// Handlers

func handleText(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	if text, ok := chunk.Message["text"]; ok {
		fmt.Fprintf(w, "%v", text)
	}
	return nil
}

func handleJSON(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	data, err := json.MarshalIndent(chunk.Message, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s\n", string(data))
	return nil
}

func handleLink(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	if text, ok := chunk.Message["text"]; ok {
		fmt.Fprintf(w, "[link] %v\n", text)
	}
	return nil
}

func handleImage(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	if url, ok := chunk.Message["url"]; ok {
		fmt.Fprintf(w, "[image] %v\n", url)
	}
	return nil
}

func handleImageLink(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	text, ok := chunk.Message["text"].(string)
	if !ok {
		return nil
	}
	if isToolFileURL(text) {
		return downloadAndPrint(text, "image", w)
	}
	fmt.Fprintf(w, "[image_link] %s\n", text)
	return nil
}

func handleFile(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	text, ok := chunk.Message["text"].(string)
	if !ok {
		if url, ok := chunk.Message["url"].(string); ok {
			text = url
		}
	}
	if text != "" && isToolFileURL(text) {
		return downloadAndPrint(text, "file", w)
	}
	fmt.Fprintf(w, "[file] %s\n", text)
	return nil
}

func handleBlob(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	mimeType, _ := chunk.Meta["mime_type"]
	fmt.Fprintf(w, "[blob] mime_type=%v\n", mimeType)
	return nil
}

func handleBlobChunk(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	if data, ok := chunk.Message["data"]; ok {
		fmt.Fprintf(w, "%v", data)
	}
	return nil
}

func handleBinaryLink(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	text, ok := chunk.Message["text"].(string)
	if !ok {
		return nil
	}
	if isToolFileURL(text) {
		return downloadAndPrint(text, "binary", w)
	}
	fmt.Fprintf(w, "[binary_link] %s\n", text)
	return nil
}

func handleVariable(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	name := chunk.Message["variable_name"]
	value := chunk.Message["variable_value"]
	if chunk.Message["stream"] == true {
		fmt.Fprintf(w, "[variable:stream] %v = %v\n", name, value)
	} else {
		fmt.Fprintf(w, "[variable] %v = %v\n", name, value)
	}
	return nil
}

func handleLog(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	label := chunk.Message["label"]
	status := chunk.Message["status"]
	id := chunk.Message["id"]
	fmt.Fprintf(w, "[log] id=%v label=%v status=%v\n", id, label, status)

	if data, ok := chunk.Message["data"]; ok && data != nil {
		if dataJSON, err := json.MarshalIndent(data, "  ", "  "); err == nil {
			fmt.Fprintf(w, "  data: %s\n", string(dataJSON))
		}
	}
	if errMsg, ok := chunk.Message["error"]; ok && errMsg != nil {
		fmt.Fprintf(w, "  error: %v\n", errMsg)
	}
	return nil
}

func handleRetrieverResources(chunk *types.DifyToolResponseChunk, w io.Writer) error {
	context := chunk.Message["context"]
	resources := chunk.Message["retriever_resources"]
	fmt.Fprintf(w, "[retriever_resources]\n")
	fmt.Fprintf(w, "  context: %v\n", context)

	if resources != nil {
		if resourcesJSON, err := json.MarshalIndent(resources, "  ", "  "); err == nil {
			fmt.Fprintf(w, "  resources: %s\n", string(resourcesJSON))
		}
	}
	return nil
}

// Download helpers

var filesBaseURL string

func SetFilesURL(url string) { filesBaseURL = strings.TrimSuffix(url, "/") }

func isToolFileURL(url string) bool {
	return strings.Contains(url, "/files/tools/")
}

func downloadAndPrint(fileURL string, fileType string, w io.Writer) error {
	url := fileURL
	if !strings.HasPrefix(fileURL, "http://") && !strings.HasPrefix(fileURL, "https://") {
		if filesBaseURL == "" {
			fmt.Fprintf(w, "[%s] %s (download skipped: files_url not configured)\n", fileType, fileURL)
			return nil
		}
		url = filesBaseURL + fileURL
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(w, "[%s] %s (download failed: %v)\n", fileType, fileURL, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "[%s] %s (download failed: status %d)\n", fileType, fileURL, resp.StatusCode)
		return nil
	}

	urlPath := fileURL
	if idx := strings.Index(urlPath, "?"); idx != -1 {
		urlPath = urlPath[:idx]
	}
	filename := filepath.Base(urlPath)

	if err := os.MkdirAll("downloads", 0755); err != nil {
		fmt.Fprintf(w, "[%s] %s (failed to create dir: %v)\n", fileType, fileURL, err)
		return nil
	}

	localPath := filepath.Join("downloads", filename)
	file, err := os.Create(localPath)
	if err != nil {
		fmt.Fprintf(w, "[%s] %s (failed to create file: %v)\n", fileType, fileURL, err)
		return nil
	}
	defer file.Close()

	if _, err = io.Copy(file, resp.Body); err != nil {
		fmt.Fprintf(w, "[%s] %s (failed to write: %v)\n", fileType, fileURL, err)
		return nil
	}

	absPath, _ := filepath.Abs(localPath)
	fmt.Fprintf(w, "[%s] downloaded to: %s\n", fileType, absPath)
	return nil
}
