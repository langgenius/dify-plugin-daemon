package serverless

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/stretchr/testify/require"
)

func TestLaunchPluginContinuesWhenRedisLockIsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/runner/instances", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{
			"error":"",
			"Items":[{
				"ID":"runtime-id",
				"Name":"runtime-name",
				"Endpoint":"https://runtime.example.test",
				"ResourceName":"runtime-resource",
				"Status":{"State":"running"}
			}]
		}`))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	previousBaseURL := baseurl
	previousClient := client
	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	baseurl = parsedURL
	client = server.Client()
	t.Cleanup(func() {
		baseurl = previousBaseURL
		client = previousClient
	})

	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	require.NoError(t, cache.Close())

	packageBytes, err := os.ReadFile("../plugin_manager/testdata/openai.difypkg")
	require.NoError(t, err)
	packageDecoder, err := decoder.NewZipPluginDecoder(packageBytes)
	require.NoError(t, err)
	identifier, err := packageDecoder.UniqueIdentity()
	require.NoError(t, err)

	response, err := LaunchPlugin(
		context.Background(),
		identifier,
		packageBytes,
		packageDecoder,
		1,
		false,
	)
	require.NoError(t, err)

	events := make([]LaunchFunctionResponse, 0, 3)
	require.NoError(t, response.Process(func(event LaunchFunctionResponse) {
		events = append(events, event)
	}))
	require.Equal(t, []LaunchFunctionResponse{
		{Event: FunctionUrl, Message: "https://runtime.example.test"},
		{Event: Function, Message: "runtime-name"},
		{Event: Done, Message: ""},
	}, events)
}
