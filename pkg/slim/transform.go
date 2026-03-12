package slim

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func TransformRequest(ctx *InvokeContext) ([]byte, string, error) {
	sessionID := uuid.New().String()

	var args map[string]any
	if err := json.Unmarshal(ctx.Request.Data, &args); err != nil {
		return nil, "", NewError(ErrInvalidArgsJSON, err.Error())
	}

	route, ok := LookupRoute(ctx.Action)
	if !ok {
		return nil, "", NewError(ErrUnknownAction, ctx.Action)
	}

	args["user_id"] = ctx.Request.UserID
	args["type"] = route.Type
	args["action"] = ctx.Action

	message := map[string]any{
		"tenant_id":       ctx.Request.TenantID,
		"session_id":      sessionID,
		"conversation_id": nil,
		"message_id":      nil,
		"app_id":          nil,
		"endpoint_id":     nil,
		"context":         map[string]any{},
		"event":           "request",
		"data":            args,
	}

	b, err := json.Marshal(message)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal message: %w", err)
	}
	return b, sessionID, nil
}
