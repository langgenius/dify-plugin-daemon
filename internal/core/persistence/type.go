package persistence

type PersistenceStorage interface {
	Save(tenant_id string, key string, data []byte) error
	Load(tenant_id string, key string) ([]byte, error)
	Delete(tenant_id string, key string) error
}
