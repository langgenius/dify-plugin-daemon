package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers/definitions"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
	"golang.org/x/tools/imports"
)

// GenerateController generates a controller file for the given access type
func GenerateController(accessType access_types.PluginAccessType, dispatchers []*definitions.PluginDispatcher) error {
	// Create template
	tmpl := template.Must(template.New("controller").Parse(controllerTemplate))

	// Create output file
	outputPath := filepath.Join("internal", "server", "controllers", strings.ToLower(string(accessType))+".gen.go")

	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, struct {
		AccessType  access_types.PluginAccessType
		Dispatchers []*definitions.PluginDispatcher
	}{
		AccessType:  accessType,
		Dispatchers: dispatchers,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Format code
	src, err := format.Source([]byte(buf.String()))
	if err != nil {
		fmt.Println(buf.String())
		return fmt.Errorf("failed to format code: %v", err)
	}

	// imports necessary packages
	output, err := imports.Process(outputPath, src, nil)
	if err != nil {
		return fmt.Errorf("failed to process imports: %v", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	// Write to file
	if _, err := f.Write(output); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// GenerateService generates a service file for the given access type
func GenerateService(accessType access_types.PluginAccessType, dispatchers []*definitions.PluginDispatcher) error {
	// Create template
	tmpl := template.Must(template.New("service").Parse(serviceTemplate))

	// Create output file
	outputPath := filepath.Join("internal", "service", strings.ToLower(string(accessType))+".gen.go")

	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, struct {
		AccessType  access_types.PluginAccessType
		Dispatchers []*definitions.PluginDispatcher
	}{
		AccessType:  accessType,
		Dispatchers: dispatchers,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Format code
	src, err := format.Source([]byte(buf.String()))
	if err != nil {
		return fmt.Errorf("failed to format code: %v", err)
	}

	// imports necessary packages
	output, err := imports.Process(outputPath, src, nil)
	if err != nil {
		return fmt.Errorf("failed to process imports: %v", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	// Write to file
	if _, err := f.Write(output); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// GeneratePluginDaemon generates a plugin daemon file for the given access type
func GeneratePluginDaemon(accessType access_types.PluginAccessType, dispatchers []*definitions.PluginDispatcher) error {
	// Create template
	tmpl := template.Must(template.New("pluginDaemon").Parse(pluginDaemonTemplate))

	// Create output file
	outputPath := filepath.Join("internal", "core", "plugin_daemon", strings.ToLower(string(accessType))+".gen.go")

	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, struct {
		AccessType  access_types.PluginAccessType
		Dispatchers []*definitions.PluginDispatcher
	}{
		AccessType:  accessType,
		Dispatchers: dispatchers,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Format code
	src, err := format.Source([]byte(buf.String()))
	if err != nil {
		return fmt.Errorf("failed to format code: %v", err)
	}

	// imports necessary packages
	output, err := imports.Process(outputPath, src, nil)
	if err != nil {
		return fmt.Errorf("failed to process imports: %v", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	// Write to file
	if _, err := f.Write(output); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// GenerateHTTPServer generates a http server file for the given access type
func GenerateHTTPServer(dispatchers []*definitions.PluginDispatcher) error {
	// Create template
	tmpl := template.Must(template.New("httpServer").Parse(httpServerTemplate))

	// Create output file
	outputPath := filepath.Join("internal", "server", "http_server.gen.go")

	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, struct {
		Dispatchers []*definitions.PluginDispatcher
	}{
		Dispatchers: dispatchers,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Format code
	src, err := format.Source([]byte(buf.String()))
	if err != nil {
		return fmt.Errorf("failed to format code: %v", err)
	}

	// imports necessary packages
	output, err := imports.Process(outputPath, src, nil)
	if err != nil {
		return fmt.Errorf("failed to process imports: %v", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	// Write to file
	if _, err := f.Write(output); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// GenerateAll generates all controller and service files based on dispatchers
func GenerateAll() error {
	// Group dispatchers by access type
	dispatchersByType := make(map[access_types.PluginAccessType][]*definitions.PluginDispatcher)
	for _, dispatcher := range definitions.PluginDispatchers {
		dispatchersByType[dispatcher.AccessType] = append(dispatchersByType[dispatcher.AccessType], &dispatcher)
	}

	// Override RequestType and ResponseType to be the actual type by using reflection
	for _, dispatchers := range dispatchersByType {
		for _, dispatcher := range dispatchers {
			dispatcher.RequestTypeString = reflect.TypeOf(dispatcher.RequestType).String()
			dispatcher.ResponseTypeString = reflect.TypeOf(dispatcher.ResponseType).String()
		}
	}

	// Generate files for each access type
	for accessType, dispatchers := range dispatchersByType {
		if err := GenerateController(accessType, dispatchers); err != nil {
			return fmt.Errorf("failed to generate controller for %s: %v", accessType, err)
		}
		if err := GenerateService(accessType, dispatchers); err != nil {
			return fmt.Errorf("failed to generate service for %s: %v", accessType, err)
		}
		if err := GeneratePluginDaemon(accessType, dispatchers); err != nil {
			return fmt.Errorf("failed to generate plugin daemon for %s: %v", accessType, err)
		}
	}

	if err := GenerateHTTPServer(
		mapping.MapArray(
			definitions.PluginDispatchers,
			func(dispatcher definitions.PluginDispatcher) *definitions.PluginDispatcher {
				return &dispatcher
			},
		),
	); err != nil {
		return fmt.Errorf("failed to generate http server: %v", err)
	}

	return nil
}
