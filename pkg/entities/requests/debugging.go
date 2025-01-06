package requests

type RequestGetRemoteDebuggingKey struct {
	TenantID string `uri:"tenant_id" validate:"required"`
}
