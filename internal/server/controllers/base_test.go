package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
)

func TestStatusCodeFromResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp *entities.Response
		want int
	}{
		{
			name: "nil response",
			resp: nil,
			want: http.StatusInternalServerError,
		},
		{
			name: "success response",
			resp: &entities.Response{Code: 0},
			want: http.StatusOK,
		},
		{
			name: "positive code response",
			resp: &entities.Response{Code: 123},
			want: http.StatusOK,
		},
		{
			name: "bad request response",
			resp: &entities.Response{Code: -400},
			want: http.StatusBadRequest,
		},
		{
			name: "not found response",
			resp: &entities.Response{Code: -404},
			want: http.StatusNotFound,
		},
		{
			name: "internal server error response",
			resp: &entities.Response{Code: -500},
			want: http.StatusInternalServerError,
		},
		{
			name: "invalid low status code",
			resp: &entities.Response{Code: -99},
			want: http.StatusInternalServerError,
		},
		{
			name: "invalid high status code",
			resp: &entities.Response{Code: -600},
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := statusCodeFromResponse(tt.resp)
			if got != tt.want {
				t.Fatalf("statusCodeFromResponse() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestJSONResponse(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		resp       *entities.Response
		wantStatus int
		wantBody   entities.Response
	}{
		{
			name:       "success response",
			resp:       entities.NewSuccessResponse(map[string]any{"ok": true}),
			wantStatus: http.StatusOK,
			wantBody: entities.Response{
				Code:    0,
				Message: "success",
				Data:    map[string]any{"ok": true},
			},
		},
		{
			name:       "bad request response",
			resp:       entities.NewDaemonErrorResponse(-400, "bad request"),
			wantStatus: http.StatusBadRequest,
			wantBody: entities.Response{
				Code:    -400,
				Message: "bad request",
				Data:    nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			JSONResponse(ctx, tt.resp)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("recorder.Code = %d, want %d", recorder.Code, tt.wantStatus)
			}

			var got entities.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if got.Code != tt.wantBody.Code {
				t.Fatalf("response code = %d, want %d", got.Code, tt.wantBody.Code)
			}
			if got.Message != tt.wantBody.Message {
				t.Fatalf("response message = %q, want %q", got.Message, tt.wantBody.Message)
			}

			if tt.wantBody.Data == nil {
				if got.Data != nil {
					t.Fatalf("response data = %#v, want nil", got.Data)
				}
				return
			}

			gotMap, ok := got.Data.(map[string]any)
			if !ok {
				t.Fatalf("response data type = %T, want map[string]any", got.Data)
			}

			wantMap := tt.wantBody.Data.(map[string]any)
			if len(gotMap) != len(wantMap) {
				t.Fatalf("response data length = %d, want %d", len(gotMap), len(wantMap))
			}

			for key, wantValue := range wantMap {
				if gotMap[key] != wantValue {
					t.Fatalf("response data[%q] = %#v, want %#v", key, gotMap[key], wantValue)
				}
			}
		})
	}
}
