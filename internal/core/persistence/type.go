package persistence

type PersistenceStorage interface {
	Save(tenant_id string, plugin_checksum string, key string, data []byte) error
	Load(tenant_id string, plugin_checksum string, key string) ([]byte, error)
	Delete(tenant_id string, plugin_checksum string, key string) error
	StateSize(tenant_id string, plugin_checksum string, key string) (int64, error)
}
