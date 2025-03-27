package cache

import (
	"errors"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
)

// Set the value with key
func AutoSet[T any](key string, value T, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	fullTypeInfo := reflect.TypeOf(value)
	pkgPath := fullTypeInfo.PkgPath()
	typeName := fullTypeInfo.Name()
	fullTypeName := pkgPath + "." + typeName

	key = serialKey("auto_type", fullTypeName, key)
	return store(key, value, time.Minute*30, context...)
}

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
	fullTypeInfo := reflect.TypeOf(result_tmpl)
	pkgPath := fullTypeInfo.PkgPath()
	typeName := fullTypeInfo.Name()
	fullTypeName := pkgPath + "." + typeName

	key = serialKey("auto_type", fullTypeName, key)
	result, err := get[T](key, context...)
	if err != nil {
		if err == ErrNotFound {
			result, err = getter()
			if err != nil {
				return nil, err
			}

			if err := store(key, result, time.Minute*30, context...); err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, err
	}

	return result, err
}

// Delete the value with key
func AutoDelete[T any](key string, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	var result_tmpl T

	fullTypeInfo := reflect.TypeOf(result_tmpl)
	pkgPath := fullTypeInfo.PkgPath()
	typeName := fullTypeInfo.Name()
	fullTypeName := pkgPath + "." + typeName

	key = serialKey("auto_type", fullTypeName, key)
	return del(key, context...)
}
