package constants

const (
	// GlobalTenantId not visible to end users.
	// Used to prevent plugins from being garbage collected when uninstalled from all user workspaces.
	GlobalTenantId = "00000000-0000-0000-0000-000000000000"
)
