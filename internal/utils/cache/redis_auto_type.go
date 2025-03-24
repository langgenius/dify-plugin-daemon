package cache

import (
	"errors"
	"reflect"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
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
	cborValue, err := parser.MarshalCBOR(value)
	if err != nil {
		return err
	}

	return getCmdable(context...).Set(ctx, key, cborValue, time.Minute*30).Err()
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
	val, err := getCmdable(context...).Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			value, err := getter()
			if err != nil {
				return nil, err
			}

			if err := store(key, value, time.Minute*30, context...); err != nil {
				return nil, err
			}
			return value, nil
		}
		return nil, err
	}

	result, err := parser.UnmarshalCBOR[T](val)
	return &result, err
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
	return getCmdable(context...).Del(ctx, key).Err()
}
