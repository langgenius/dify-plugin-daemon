package io_tunnel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

func getBasicPluginAccessMap(
	user_id string,
	access_type access_types.PluginAccessType,
	action access_types.PluginAccessAction,
) map[string]any {
	return map[string]any{
		"user_id": user_id,
		"type":    access_type,
		"action":  action,
	}
}

func GetInvokePluginMap(
	session *session_manager.Session,
	request any,
) map[string]any {
	req := getBasicPluginAccessMap(
		session.UserID,
		session.InvokeFrom,
		session.Action,
	)
	for k, v := range parser.StructToMap(request) {
		req[k] = v
	}
	return req
}
