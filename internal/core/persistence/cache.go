package persistence

type Persistence struct {
	storage PersistenceStorage
}

func (c *Persistence) Save(tenant_id string, key string, data []byte) error {
	return nil
}

func (c *Persistence) Load(tenant_id string, key string) ([]byte, error) {
	return nil, nil
}

func (c *Persistence) Delete(tenant_id string, key string) error {
	return nil
}

func (c *Persistence) Scan(tenant_id string, prefix string, cursor int64) ([]string, error) {
	return nil, nil
}
