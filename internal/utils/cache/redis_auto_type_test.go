package cache

import (
	"errors"
	"testing"
)

type TestAutoTypeStruct struct {
	ID string `json:"id"`
}

func TestAutoType(t *testing.T) {
	if err := InitRedisClient("127.0.0.1:6379", "difyai123456", false, 0); err != nil {
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

	if err := AutoDelete[TestAutoTypeStruct]("test"); err != nil {
		t.Fatal(err)
	}
}

func TestAutoTypeWithGetter(t *testing.T) {
	if err := InitRedisClient("127.0.0.1:6379", "difyai123456", false, 0); err != nil {
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
	if err := AutoDelete[TestAutoTypeStruct]("test1"); err != nil {
		t.Fatal(err)
	}

	if result.ID != "123" {
		t.Fatal("result not correct")
	}
}
