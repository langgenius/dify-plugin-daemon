package requests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

func TestValidateToolCredentialsRequest(t *testing.T) {
	for _, tt := range []struct {
		name           string
		credentialType string
		resultFormat   string
		valid          bool
	}{
		{name: "legacy", valid: true},
		{name: "api key", credentialType: "api_key", resultFormat: "detailed", valid: true},
		{name: "oauth", credentialType: "oauth2", resultFormat: "detailed", valid: true},
		{name: "missing credential type", resultFormat: "detailed"},
		{name: "unknown credential type", credentialType: "unknown", resultFormat: "detailed"},
		{name: "unknown format", credentialType: "api_key", resultFormat: "unknown"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := RequestValidateToolCredentials{
				Provider: "calendar", Credentials: map[string]any{"token": "test-token"},
				CredentialType: tt.credentialType, ResultFormat: tt.resultFormat,
			}
			err := validators.GlobalEntitiesValidator.Struct(request)
			if !tt.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			payload := parser.StructToMap(request)
			if tt.resultFormat == "" {
				require.NotContains(t, payload, "result_format")
				require.NotContains(t, payload, "credential_type")
			} else {
				require.Equal(t, tt.resultFormat, payload["result_format"])
				require.Equal(t, tt.credentialType, payload["credential_type"])
			}
		})
	}
}
