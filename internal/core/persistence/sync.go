package persistence

// Sync sync the cache to storage
// NOTE: this method is not used currently, for now, we assume that
// the effective is fast enough, but maybe someday we need to use the cache
// func (p *Persistence) Sync() error {
// 	// sync cache to storage
// 	cache.ScanKeysAsync(fmt.Sprintf("%s:*", CACHE_KEY_PREFIX), func(keys []string) error {
// 	})

// 	return nil
// }
