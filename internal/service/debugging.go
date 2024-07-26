package service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/remote_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

func GetRemoteDebuggingKey(tenant_id string) entities.Response {
	type response struct {
		Key string `json:"key"`
	}

	key, err := remote_manager.GetConnectionKey(remote_manager.ConnectionInfo{
		TenantId: tenant_id,
	})

	if err != nil {
		return *entities.NewErrorResponse(-500, err.Error())
	}

	return *entities.NewSuccessResponse(response{
		Key: key,
	})
}
