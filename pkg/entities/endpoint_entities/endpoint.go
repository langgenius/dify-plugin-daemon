package endpoint_entities

type SetupEndpointResponse struct {
	Success bool `json:"success" validate:"required"`
}

type EndpointResponseChunk struct {
	Status  *uint16           `json:"status" validate:"omitempty"`
	Headers map[string]string `json:"headers" validate:"omitempty"`
	Result  *string           `json:"result" validate:"omitempty"`
}

const (
	HeaderXOriginalHost = "X-Original-Host"
)
