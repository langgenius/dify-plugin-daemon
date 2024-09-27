package cache

import "testing"

type TestAutoTypeStruct struct {
	ID string `json:"id"`
}

func TestAutoType(t *testing.T) {
	if err := InitRedisClient("127.0.0.1:6379", "difyai123456"); err != nil {
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
}
