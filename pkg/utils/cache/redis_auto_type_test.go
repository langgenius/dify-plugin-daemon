package cache

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestAutoTypeStruct struct {
	ID string `json:"id"`
}

func TestAutoType(t *testing.T) {
	if err := InitRedisClient("127.0.0.1:6379", "", "difyai123456", false, 0, nil); err != nil {
		t.Fatal(err)
	}
	defer Close()

	err := AutoSet("test", TestAutoTypeStruct{ID: "123"})
	if err != nil {
		t.Fatal(err)
	}

	result, err := AutoGet[TestAutoTypeStruct]("test")
	if err != nil {
		t.Fatal(err)
	}

	if result.ID != "123" {
		t.Fatal("result not correct")
	}

	if _, err := AutoDelete[TestAutoTypeStruct]("test"); err != nil {
		t.Fatal(err)
	}
}

func TestAutoTypeWithGetter(t *testing.T) {
	if err := InitRedisClient("127.0.0.1:6379", "", "difyai123456", false, 0, nil); err != nil {
		t.Fatal(err)
	}
	defer Close()

	result, err := AutoGetWithGetter("test1", func() (*TestAutoTypeStruct, error) {
		return &TestAutoTypeStruct{
			ID: "123",
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err = AutoGetWithGetter("test1", func() (*TestAutoTypeStruct, error) {
		return nil, errors.New("must hit cache")
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AutoDelete[TestAutoTypeStruct]("test1"); err != nil {
		t.Fatal(err)
	}

	if result.ID != "123" {
		t.Fatal("result not correct")
	}
}

func TestAutoTypeUsesConfiguredPrefix(t *testing.T) {
	if err := InitRedisClient("127.0.0.1:6379", "", "difyai123456", false, 0, nil); err != nil {
		t.Fatal(err)
	}
	defer Close()

	SetKeyPrefix("enterprise-a")
	t.Cleanup(func() { SetKeyPrefix("plugin_daemon") })

	require.NoError(t, AutoSet("typed:key", TestAutoTypeStruct{ID: "123"}))

	fullTypeInfo := reflect.TypeOf(TestAutoTypeStruct{})
	fullTypeName := fullTypeInfo.PkgPath() + "." + fullTypeInfo.Name()
	physicalKey := "enterprise-a:auto_type:" + fullTypeName + ":typed:key"
	t.Cleanup(func() {
		_ = client.Del(ctx, physicalKey).Err()
	})

	result, err := client.Get(ctx, physicalKey).Bytes()
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	value, err := AutoGetWithGetter("typed:key", func() (*TestAutoTypeStruct, error) {
		return nil, errors.New("must hit cache")
	})
	require.NoError(t, err)
	assert.Equal(t, "123", value.ID)

	_, err = AutoDelete[TestAutoTypeStruct]("typed:key")
	require.NoError(t, err)

	_, err = client.Get(ctx, physicalKey).Result()
	assert.Error(t, err)
}
