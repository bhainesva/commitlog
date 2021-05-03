package commitlog

import (
	"testing"
)

func TestCacheWriteThenRead(t *testing.T) {
	ch := make(chan cacheRequest)
	go initializeCache(ch)
	expectedValue := "test"
	key := "key"

	ch <- cacheRequest{
		Type:    WRITE,
		Payload: expectedValue,
		Key: key,
	}

	out := make(chan cacheRequest)
	ch <- cacheRequest{
		Type:    READ,
		Key:     key,
		Out: out,
	}

	got := <-out
	gotStr, ok := got.Payload.(string)
	if !ok {
		t.Errorf("Expected to find string value on cache read, got: %#v", got.Payload)
	}
	if gotStr != expectedValue {
		t.Errorf("Read %s from cache; expected %s", gotStr, expectedValue)
	}
}

func TestCacheWriteHelper(t *testing.T) {
	ch := make(chan cacheRequest)
	go initializeCache(ch)
	key := "key"
	expectedValue := "test"

	writeCacheEntry(ch, key, expectedValue)

	out := make(chan cacheRequest)
	ch <- cacheRequest{
		Type:    READ,
		Key:    key,
		Out: out,
	}

	got := <-out
	gotStr, ok := got.Payload.(string)
	if !ok {
		t.Errorf("Expected to find string value on cache read, got: %#v", got.Payload)
	}
	if gotStr != expectedValue {
		t.Errorf("Read %s from cache; expected %s", gotStr, expectedValue)
	}
}

func TestCacheReadEmpty(t *testing.T) {
	ch := make(chan cacheRequest)
	go initializeCache(ch)

	out := make(chan cacheRequest)
	ch <- cacheRequest{
		Type:    READ,
		Key:     "key",
		Out: out,
	}

	got := <-out
	if got.Payload != nil {
		t.Errorf("Expected nil payload when reading non-existent key")
	}
}

func TestCacheDelete(t *testing.T) {
	ch := make(chan cacheRequest)
	go initializeCache(ch)
	key := "key"

	ch <- cacheRequest{
		Type:    WRITE,
		Payload: "value",
		Key:     key,
	}

	ch <- cacheRequest{
		Type:    DELETE,
		Key:     key,
	}

	out := make(chan cacheRequest)
	ch <- cacheRequest{
		Type:    READ,
		Key:     key,
		Out: out,
	}

	got := <-out
	if got.Payload != nil {
		t.Errorf("Expected nil payload when reading deleted key")
	}
}
