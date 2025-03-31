package cache

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TEST_PREFIX = "test"
)

func getRedisConnection() error {
	return InitRedisClient("0.0.0.0:6379", "difyai123456", false, 0)
}

func TestRedisConnection(t *testing.T) {
	// get redis connection
	if err := getRedisConnection(); err != nil {
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
	if err := getRedisConnection(); err != nil {
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

func TestRedisScanMap(t *testing.T) {
	// get redis connection
	if err := getRedisConnection(); err != nil {
		t.Errorf("get redis connection failed: %v", err)
		return
	}
	defer Close()

	type s struct {
		Field string `json:"field"`
	}

	err := SetMapOneField(strings.Join([]string{TEST_PREFIX, "map"}, ":"), "key1", s{Field: "value1"})
	if err != nil {
		t.Errorf("set map failed: %v", err)
		return
	}
	defer Del(strings.Join([]string{TEST_PREFIX, "map"}, ":"))
	err = SetMapOneField(strings.Join([]string{TEST_PREFIX, "map"}, ":"), "key2", s{Field: "value2"})
	if err != nil {
		t.Errorf("set map failed: %v", err)
		return
	}
	err = SetMapOneField(strings.Join([]string{TEST_PREFIX, "map"}, ":"), "key3", s{Field: "value3"})
	if err != nil {
		t.Errorf("set map failed: %v", err)
		return
	}
	err = SetMapOneField(strings.Join([]string{TEST_PREFIX, "map"}, ":"), "4", s{Field: "value4"})
	if err != nil {
		t.Errorf("set map failed: %v", err)
		return
	}

	data, err := ScanMap[s](strings.Join([]string{TEST_PREFIX, "map"}, ":"), "key*")
	if err != nil {
		t.Errorf("scan map failed: %v", err)
		return
	}

	if len(data) != 3 {
		t.Errorf("scan map should return 3")
		return
	}

	if data["key1"].Field != "value1" {
		t.Errorf("scan map should return value1")
		return
	}

	if data["key2"].Field != "value2" {
		t.Errorf("scan map should return value2")
		return
	}

	if data["key3"].Field != "value3" {
		t.Errorf("scan map should return value3")
		return
	}

	err = ScanMapAsync[s](strings.Join([]string{TEST_PREFIX, "map"}, ":"), "4", func(m map[string]s) error {
		if len(m) != 1 {
			t.Errorf("scan map async should return 1")
			return errors.New("scan map async should return 1")
		}

		if m["4"].Field != "value4" {
			t.Errorf("scan map async should return value4")
			return errors.New("scan map async should return value4")
		}

		return nil
	})

	if err != nil {
		t.Errorf("scan map async failed: %v", err)
		return
	}
}

func TestRedisP2PPubsub(t *testing.T) {
	// get redis connection
	if err := getRedisConnection(); err != nil {
		t.Errorf("get redis connection failed: %v", err)
		return
	}
	defer Close()

	ch := "test-channel"

	type s struct{}

	sub, cancel := Subscribe[s](ch)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		<-sub
		wg.Done()
	}()

	// test pubsub
	err := Publish(ch, s{})
	if err != nil {
		t.Errorf("publish failed: %v", err)
		return
	}

	wg.Wait()
}

func TestRedisP2ARedis(t *testing.T) {
	// get redis connection
	if err := getRedisConnection(); err != nil {
		t.Errorf("get redis connection failed: %v", err)
		return
	}
	defer Close()

	ch := "test-channel-p2a"

	type s struct{}

	wg := sync.WaitGroup{}
	wg.Add(3)

	swg := sync.WaitGroup{}
	swg.Add(3)

	for i := 0; i < 3; i++ {
		go func() {
			sub, cancel := Subscribe[s](ch)
			swg.Done()
			defer cancel()
			<-sub
			wg.Done()
		}()
	}

	swg.Wait()

	// test pubsub
	err := Publish(ch, s{})
	if err != nil {
		t.Errorf("publish failed: %v", err)
		return
	}

	wg.Wait()
}

func TestGetRedisOptions(t *testing.T) {
	opts := getRedisOptions("dummy:6379", "password", false, 0)
	if opts.TLSConfig != nil {
		t.Errorf("TLSConfig should not be set")
		return
	}

	opts = getRedisOptions("dummy:6379", "password", true, 0)
	if opts.TLSConfig == nil {
		t.Errorf("TLSConfig should be set")
		return
	}
}

func TestSetAndGet(t *testing.T) {
	if err := InitRedisClient("127.0.0.1:6379", "difyai123456", false, 0); err != nil {
		t.Fatal(err)
	}
	defer Close()

	m := map[string]string{
		"key": "hello",
	}

	err := Store(strings.Join([]string{TEST_PREFIX, "get-test"}, ":"), m, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	val, err := Get[map[string]string](strings.Join([]string{TEST_PREFIX, "get-test"}, ":"))
	if err != nil {
		t.Fatal(err)
	}
	if (*val)["key"] != "hello" {
		t.Fatalf("Get[\"key\"] should be \"hello\"")
	}
	err = Del(strings.Join([]string{TEST_PREFIX, "get-test"}, ":"))
	if err != nil {
		t.Fatal(err)
	}
	val, err = Get[map[string]string](strings.Join([]string{TEST_PREFIX, "get-test"}, ":"))
	if err != ErrNotFound {
		t.Fatalf("Get[\"key\"] should be ErrNotFound")
	}
}
