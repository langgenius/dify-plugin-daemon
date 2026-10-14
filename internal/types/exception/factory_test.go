package exception

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalServerErrorRedactsResponseAndLogsDetails(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	err := errors.New("cannot open /private/example with token=test-secret")
	response := InternalServerError(err).ToResponse()
	data, marshalErr := json.Marshal(response)
	require.NoError(t, marshalErr)
	require.JSONEq(t, `{"code":-500,"message":"{\"message\":\"Internal Server Error\",\"error_type\":\"PluginDaemonInternalServerError\",\"args\":null}","data":null}`, string(data))
	require.NotContains(t, string(data), "test-secret")
	require.Contains(t, logs.String(), err.Error())
	require.Contains(t, logs.String(), `"traceback":`)
	require.Contains(t, logs.String(), "TestInternalServerErrorRedactsResponseAndLogsDetails")
}
