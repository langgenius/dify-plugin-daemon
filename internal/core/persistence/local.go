package persistence

type LocalWrapper struct{}

func NewLocalWrapper() *LocalWrapper {
	return &LocalWrapper{}
}

func (l *LocalWrapper) Save(tenant_id string, key string, data []byte) error {
	return nil
}

func (l *LocalWrapper) Load(tenant_id string, key string) ([]byte, error) {
	return nil, nil
}

func (l *LocalWrapper) Delete(tenant_id string, key string) error {
	return nil
}
