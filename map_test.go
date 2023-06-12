package orderMap_test

import (
	orderMap "github.com/RealFax/order-map"
	"strconv"
	"testing"
)

var (
	testMap = orderMap.New[string, string]()
)

func init() {
	testMap.Store("hello", "bonjour")
}

func TestMap_Store(t *testing.T) {
	testMap.Store("hello1", "bonjour")
}

func TestMap_Load(t *testing.T) {
	val, ok := testMap.Load("hello")
	t.Log("State:", ok, ", Value:", val)
}

func TestMap_Delete(t *testing.T) {
	testMap.Store("hello1", "bonjour")
	val, ok := testMap.Load("hello1")
	t.Log("State:", ok, ", Value:", val)

	testMap.Delete("hello1")
	val, ok = testMap.Load("hello1")
	t.Log("State:", ok, ", Value:", val)
}

func TestMap_Range(t *testing.T) {
	for i := 0; i < 5; i++ {
		testMap.Store("test"+strconv.Itoa(i), "range!")
	}

	testMap.Range(func(key string, value string) bool {
		t.Log("Key:", key, ", Value:", value)
		return true
	})
}

func TestMap_DisorderedRange(t *testing.T) {
	for i := 0; i < 5; i++ {
		testMap.Store("_test"+strconv.Itoa(i), "disordered_range!")
	}

	testMap.DisorderedRange(func(key string, value string) bool {
		t.Log("Key:", key, ", Value:", value)
		return true
	})
}

func BenchmarkMap_Store(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testMap.Store("key", "value")
	}
}

func BenchmarkMap_Load(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testMap.Load("hello")
	}
}
