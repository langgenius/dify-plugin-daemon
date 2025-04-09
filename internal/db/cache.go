package db

import (
	"fmt"
	"reflect"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

const (
	CACHE_PREFIX      = "cache"
	CACHE_EXPIRE_TIME = time.Minute * 5
)

type KeyValuePair struct {
	Key string
	Val any
}

type GetCachePayload[T any] struct {
	Getter   func() (*T, error)
	CacheKey []KeyValuePair
}

func joinCacheKey(typename string, pairs []KeyValuePair) string {
	cacheKey := CACHE_PREFIX + ":" + typename
	for _, kv := range pairs {
		cacheKey += ":" + kv.Key + ":"
		// convert value to string
		cacheKey += fmt.Sprintf("%v", kv.Val)
	}
	return cacheKey
}

func GetCache[T any](p *GetCachePayload[T]) (*T, error) {
	var t T
	typename := reflect.TypeOf(t).String()

	// join cache key
	cacheKey := joinCacheKey(typename, p.CacheKey)

	// get cache
	val, err := cache.Get[T](cacheKey)
	if err == nil {
		return val, nil
	}

	if err == cache.ErrNotFound {
		// get from getter
		val, err := p.Getter()
		if err != nil {
			return nil, err
		}

		// set cache
		err = cache.Store(cacheKey, val, CACHE_EXPIRE_TIME)
		if err != nil {
			return nil, err
		}

		return val, nil
	} else {
		return nil, err
	}
}

type DeleteCachePayload[T any] struct {
	Delete   func() error
	CacheKey []KeyValuePair
}

func DeleteCache[T any](p *DeleteCachePayload[T]) error {
	var t T
	typename := reflect.TypeOf(t).String()

	// join cache key
	cacheKey := joinCacheKey(typename, p.CacheKey)

	// delete cache
	_, err := cache.Del(cacheKey)
	if err != nil {
		return err
	}

	err = p.Delete()
	if err != nil {
		return err
	}

	return nil
}

type UpdateCachePayload[T any] struct {
	Update   func() error
	CacheKey []KeyValuePair
}

func UpdateCache[T any](p *UpdateCachePayload[T]) error {
	var t T
	typename := reflect.TypeOf(t).String()

	// join cache key
	cacheKey := joinCacheKey(typename, p.CacheKey)

	err := p.Update()
	if err != nil {
		return err
	}

	// delete cache
	_, err = cache.Del(cacheKey)
	if err != nil {
		return err
	}

	return nil
}
