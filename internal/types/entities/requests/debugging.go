package requests

type RequestGetRemoteDebuggingKey struct {
	TenantID string `json:"tenant_id" validate:"required"`
}
