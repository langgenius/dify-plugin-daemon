package cache

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TEST_PREFIX = "test"
)

func getRedisConnection(t *testing.T) error {
	return InitRedisClient("0.0.0.0:6379", "difyai123456")
}

func TestRedisConnection(t *testing.T) {
	// get redis connection
	if err := getRedisConnection(t); err != nil {
		t.Errorf("get redis connection failed: %v", err)
		return
	}

	// close
	if err := Close(); err != nil {
		t.Errorf("close redis client failed: %v", err)
		return
	}
}

func TestRedisTransaction(t *testing.T) {
	// get redis connection
	if err := getRedisConnection(t); err != nil {
		t.Errorf("get redis connection failed: %v", err)
		return
	}
	defer Close()

	// test transaction
	err := Transaction(func(p redis.Pipeliner) error {
		// set key
		if err := Store(
			strings.Join([]string{TEST_PREFIX, "key"}, ":"),
			"value",
			time.Second,
			p,
		); err != nil {
			t.Errorf("store key failed: %v", err)
			return err
		}

		return errors.New("test transaction error")
	})

	if err == nil {
		t.Errorf("transaction should return error")
		return
	}

	// get key
	value, err := GetString(
		strings.Join([]string{TEST_PREFIX, "key"}, ":"),
	)

	if err != ErrNotFound {
		t.Errorf("key should not exist")
		return
	}

	if value != "" {
		t.Errorf("value should be empty")
		return
	}

	// test success transaction
	err = Transaction(func(p redis.Pipeliner) error {
		// set key
		if err := Store(
			strings.Join([]string{TEST_PREFIX, "key"}, ":"),
			"value",
			time.Second,
			p,
		); err != nil {
			t.Errorf("store key failed: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		t.Errorf("transaction should not return error")
		return
	}

	defer Del(strings.Join([]string{TEST_PREFIX, "key"}, ":"))

	// get key
	value, err = GetString(
		strings.Join([]string{TEST_PREFIX, "key"}, ":"),
	)

	if err != nil {
		t.Errorf("get key failed: %v", err)
		return
	}

	if value != "value" {
		t.Errorf("value should be value")
		return
	}
}
