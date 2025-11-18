package service

// NOTE: ENTERPRISE ONLY

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
)

func joinGlobalTenantIfNeeded(config *app.Config, tenantId string) []string {
	tenants := []string{tenantId}
	if tenantId != constants.GlobalTenantId && config.PluginAllowOrphans {
		tenants = append(tenants, constants.GlobalTenantId)
	}
	return tenants
}
