package cache

import (
	"errors"
	"reflect"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/redis/go-redis/v9"
)

// Get the value with key
func AutoGet[T any](key string, context ...redis.Cmdable) (*T, error) {
	return AutoGetWithGetter(key, func() (*T, error) {
		return nil, errors.New("not found")
	}, context...)
}

// Get the value with key, fallback to getter if not found, and set the value to cache
func AutoGetWithGetter[T any](key string, getter func() (*T, error), context ...redis.Cmdable) (*T, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	var result_tmpl T

	// fetch full type info
	full_type_info := reflect.TypeOf(result_tmpl)
	full_type_name := full_type_info.Name()

	key = serialKey("auto_type", full_type_name, key)
	val, err := getCmdable(context...).Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			value, err := getter()
			if err != nil {
				return nil, err
			}

			if err := Store(key, value, 0, context...); err != nil {
				return nil, err
			}
			return value, nil
		}
		return nil, err
	}

	result, err := parser.UnmarshalJson[T](val)
	return &result, err
}
