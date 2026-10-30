package controlpanel

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

type countingPackageStorage struct {
	cloudoss.OSS
	loads    atomic.Int32
	data     []byte
	packages map[string][]byte
}

func (s *countingPackageStorage) Load(name string) ([]byte, error) {
	s.loads.Add(1)
	if s.packages != nil {
		return append([]byte(nil), s.packages[name]...), nil
	}
	if s.data == nil {
		return nil, errors.New("package storage must not be read")
	}
	return append([]byte(nil), s.data...), nil
}

const existingFunctionJSON = `{"error":"","Items":[{"ID":"id","Name":"runtime","Endpoint":"https://runtime.example.test","ResourceName":"resource","Status":{"State":"running"}}]}`

func serverlessPanel(t *testing.T, storage *countingPackageStorage, handler http.HandlerFunc) *ControlPanel {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/ping" {
			_, _ = w.Write([]byte(`"pong"`))
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	key := "test-key"
	config := &app.Config{
		DifyPluginServerlessConnectorURL:           &server.URL,
		DifyPluginServerlessConnectorAPIKey:        &key,
		DifyPluginServerlessConnectorLaunchTimeout: 5,
		PluginInstallTimeout:                       1,
	}
	serverless.Init(config)
	return NewControlPanel(config, nil, media_transport.NewPackageBucket(storage, "packages"), nil, nil)
}

func TestInstallExistingServerlessRuntimeDoesNotReadPackage(t *testing.T) {
	storage := &countingPackageStorage{}
	var queries atomic.Int32
	panel := serverlessPanel(t, storage, func(w http.ResponseWriter, r *http.Request) {
		queries.Add(1)
		if r.URL.Path != "/v1/runner/instances" || r.URL.Query().Get("filename") != "author@plugin@1.0.0@0123456789abcdef0123456789abcdef.difypkg" {
			t.Errorf("unexpected request: %s", r.URL)
		}
		_, _ = w.Write([]byte(existingFunctionJSON))
	})
	identifier := plugin_entities.PluginUniqueIdentifier("author/plugin:1.0.0@0123456789abcdef0123456789abcdef")
	response, err := panel.InstallToServerless(context.Background(), identifier)
	require.NoError(t, err)
	var events []serverless.LaunchFunctionResponse
	require.NoError(t, response.Process(func(event serverless.LaunchFunctionResponse) { events = append(events, event) }))
	require.Equal(t, []serverless.LaunchFunctionResponse{
		{Event: serverless.FunctionUrl, Message: "https://runtime.example.test"},
		{Event: serverless.Function, Message: "runtime"},
		{Event: serverless.Done},
	}, events)
	require.EqualValues(t, 0, storage.loads.Load())
	require.EqualValues(t, 1, queries.Load())
}

// Observe when the caller reaches its cancellation-aware wait. This keeps all
// callers overlapping without depending on sleeps or scheduler timing.
type observedInstallContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *observedInstallContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestConcurrentExistingServerlessInstallsShareQuery(t *testing.T) {
	storage := &countingPackageStorage{}
	release := make(chan struct{})
	var queries atomic.Int32
	panel := serverlessPanel(t, storage, func(w http.ResponseWriter, r *http.Request) {
		queries.Add(1)
		<-release
		_, _ = w.Write([]byte(existingFunctionJSON))
	})
	// Unblock the server even if an assertion fails.
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	const workers = 32
	results := make(chan error, workers)
	for range workers {
		ctx := &observedInstallContext{Context: t.Context(), waiting: make(chan struct{})}
		go func() {
			response, err := panel.InstallToServerless(ctx, "author/plugin:1.0.0@0123456789abcdef0123456789abcdef")
			if err == nil {
				var done bool
				err = response.Process(func(e serverless.LaunchFunctionResponse) { done = done || e.Event == serverless.Done })
				if err == nil && !done {
					err = errors.New("missing done event")
				}
			}
			results <- err
		}()
		select {
		case <-ctx.waiting:
		case <-time.After(5 * time.Second):
			t.Fatal("caller did not start waiting")
		}
	}
	releaseOnce.Do(func() { close(release) })
	for range workers {
		require.NoError(t, <-results)
	}
	require.EqualValues(t, 1, queries.Load())
	require.EqualValues(t, 0, storage.loads.Load())
}

func loadServerlessFixture(t *testing.T) ([]byte, plugin_entities.PluginUniqueIdentifier) {
	t.Helper()
	data, err := os.ReadFile("../plugin_manager/testdata/openai.difypkg")
	require.NoError(t, err)
	d, err := decoder.NewZipPluginDecoder(data)
	require.NoError(t, err)
	id, err := d.UniqueIdentity()
	require.NoError(t, err)
	return data, id
}

func writeLaunchSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	_, _ = fmt.Fprint(w, "data: {\"Stage\":\"run\",\"State\":\"success\",\"Message\":\"endpoint=https://runtime.example.test,name=runtime,id=id\"}\n\n")
	_, _ = fmt.Fprint(w, "data: {\"Stage\":\"end\",\"State\":\"success\"}\n\n")
}

func TestConcurrentMissingServerlessInstallsReadAndLaunchOnce(t *testing.T) {
	routine.InitPool(128)
	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() { _ = cache.Close() })
	data, id := loadServerlessFixture(t)
	storage := &countingPackageStorage{data: data}
	var queries, launches atomic.Int32
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	panel := serverlessPanel(t, storage, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/runner/instances" {
			queries.Add(1)
			_, _ = w.Write([]byte(`{"Items":[]}`))
			return
		}
		launches.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		<-release
		writeLaunchSuccess(w)
	})
	const workers = 32
	results := make(chan error, workers)
	for range workers {
		ctx := &observedInstallContext{Context: t.Context(), waiting: make(chan struct{})}
		go func() {
			response, err := panel.InstallToServerless(ctx, id)
			if err == nil {
				err = response.Process(func(serverless.LaunchFunctionResponse) {})
			}
			results <- err
		}()
		select {
		case <-ctx.waiting:
		case <-time.After(5 * time.Second):
			t.Fatal("caller did not start waiting")
		}
	}
	once.Do(func() { close(release) })
	for range workers {
		require.NoError(t, <-results)
	}
	require.EqualValues(t, 1, storage.loads.Load())
	require.EqualValues(t, 1, launches.Load())
	require.EqualValues(t, 2, queries.Load(), "query before and after acquiring lock")
}

func TestCancelServerlessWaiterDoesNotCancelOtherTenants(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	var queries atomic.Int32
	panel := serverlessPanel(t, &countingPackageStorage{}, func(w http.ResponseWriter, r *http.Request) {
		queries.Add(1)
		<-release
		_, _ = w.Write([]byte(existingFunctionJSON))
	})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	first := &observedInstallContext{Context: ctx, waiting: make(chan struct{})}
	second := &observedInstallContext{Context: t.Context(), waiting: make(chan struct{})}
	id := plugin_entities.PluginUniqueIdentifier("author/plugin:1.0.0@0123456789abcdef0123456789abcdef")
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go func() { _, err := panel.InstallToServerless(first, id); firstResult <- err }()
	<-first.waiting
	go func() { _, err := panel.InstallToServerless(second, id); secondResult <- err }()
	<-second.waiting
	cancel()
	select {
	case err := <-firstResult:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("canceled tenant is still waiting")
	}
	once.Do(func() { close(release) })
	require.NoError(t, <-secondResult)
	require.EqualValues(t, 1, queries.Load())
}

func TestServerlessLookupErrorDoesNotReadOrLaunch(t *testing.T) {
	storage := &countingPackageStorage{}
	panel := serverlessPanel(t, storage, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/runner/instances", r.URL.Path)
		_, _ = w.Write([]byte(`{"error":"connector unavailable"}`))
	})
	_, err := panel.InstallToServerless(t.Context(), "author/plugin:1.0.0@0123456789abcdef0123456789abcdef")
	require.ErrorContains(t, err, "connector unavailable")
	require.Zero(t, storage.loads.Load())
}

func TestServerlessLaunchFailureCannotBecomeSuccess(t *testing.T) {
	routine.InitPool(128)
	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() { _ = cache.Close() })
	data, id := loadServerlessFixture(t)
	storage := &countingPackageStorage{data: data}
	panel := serverlessPanel(t, storage, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/runner/instances" {
			_, _ = w.Write([]byte(`{"Items":[]}`))
			return
		}
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"Stage\":\"build\",\"State\":\"failed\",\"Message\":\"image missing\"}\n\n")
		// A trailing end event must never erase an earlier failure.
		writeLaunchSuccess(w)
	})
	response, err := panel.InstallToServerless(t.Context(), id)
	require.ErrorContains(t, err, "image missing")
	require.Nil(t, response)
	// Failure is not cached: a later attempt can retry the launch.
	_, err = panel.InstallToServerless(t.Context(), id)
	require.ErrorContains(t, err, "image missing")
	require.EqualValues(t, 2, storage.loads.Load())
}

func TestServerlessInstallsKeepExactIdentifiers(t *testing.T) {
	routine.InitPool(128)
	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() { _ = cache.Close() })
	firstData, firstID := loadServerlessFixture(t)
	// Same author/name/version, different package checksum: these must not share a result.
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	reader, err := zip.NewReader(bytes.NewReader(firstData), int64(len(firstData)))
	require.NoError(t, err)
	for _, file := range reader.File {
		require.NoError(t, writer.Copy(file))
	}
	extra, err := writer.Create("concurrency-test.txt")
	require.NoError(t, err)
	_, err = extra.Write([]byte("different package"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	secondData := buffer.Bytes()
	d, err := decoder.NewZipPluginDecoder(secondData)
	require.NoError(t, err)
	secondID, err := d.UniqueIdentity()
	require.NoError(t, err)
	require.NotEqual(t, firstID, secondID)
	storage := &countingPackageStorage{packages: map[string][]byte{
		path.Join("packages", firstID.String()):  firstData,
		path.Join("packages", secondID.String()): secondData,
	}}
	launched := make(chan struct{}, 2)
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	var queries atomic.Int32
	secondRechecked := make(chan struct{})
	panel := serverlessPanel(t, storage, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/runner/instances" {
			if queries.Add(1) == 4 {
				close(secondRechecked)
			}
			_, _ = w.Write([]byte(`{"Items":[]}`))
			return
		}
		_, _ = io.Copy(io.Discard, r.Body)
		launched <- struct{}{}
		<-release
		writeLaunchSuccess(w)
	})
	results := make(chan error, 2)
	go func() { _, err := panel.InstallToServerless(t.Context(), firstID); results <- err }()
	<-launched
	go func() { _, err := panel.InstallToServerless(t.Context(), secondID); results <- err }()
	<-secondRechecked
	// Both distinct packages must launch; there is no global concurrency policy here.
	select {
	case <-launched:
	case <-time.After(5 * time.Second):
		t.Fatal("different checksums incorrectly shared a launch")
	}

	once.Do(func() { close(release) })
	require.NoError(t, <-results)
	require.NoError(t, <-results)
	require.EqualValues(t, 2, storage.loads.Load())
}
