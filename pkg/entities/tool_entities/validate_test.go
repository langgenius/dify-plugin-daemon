package tool_entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

func TestValidateCredentialsResultPreservesWireFormat(t *testing.T) {
	for _, payload := range []string{
		`{"result":true}`,
		`{"result":false}`,
		`{"status":"valid","reason_code":null,"retryable":false}`,
		`{"status":"invalid","reason_code":"credential_revoked","retryable":false}`,
		`{"status":"temporarily_unavailable","reason_code":"provider_timeout","retryable":true}`,
		`{"status":"unsupported","reason_code":null,"retryable":false}`,
	} {
		t.Run(payload, func(t *testing.T) {
			result, err := parser.UnmarshalJson[ValidateCredentialsResult](payload)
			require.NoError(t, err)
			encoded, err := json.Marshal(result)
			require.NoError(t, err)
			require.JSONEq(t, payload, string(encoded))
		})
	}
}

func TestValidateCredentialsResultRejectsInvalidResponses(t *testing.T) {
	for _, payload := range []string{
		`{}`,
		`{"result":null}`,
		`{"result":true,"status":"valid"}`,
		`{"status":"unknown"}`,
	} {
		t.Run(payload, func(t *testing.T) {
			_, err := parser.UnmarshalJson[ValidateCredentialsResult](payload)
			require.Error(t, err)
		})
	}
}
