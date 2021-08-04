package cache

import (
	"testing"
)

func TestCacheWriteThenRead(t *testing.T) {
	ch := New()
	expectedValue := "test"
	key := "key"

	ch.Write(key, expectedValue)
	got := ch.Read(key)
	gotStr, ok := got.(string)
	if !ok {
		t.Errorf("Expected to find string value on cache read, got: %#v", got)
	}
	if gotStr != expectedValue {
		t.Errorf("Read %s from cache; expected %s", gotStr, expectedValue)
	}
}

func TestCacheWriteHelper(t *testing.T) {
	ch := New()
	key := "key"
	expectedValue := "test"

	ch.Write(key, expectedValue)

	got := ch.Read(key)
	gotStr, ok := got.(string)
	if !ok {
		t.Errorf("Expected to find string value on cache read, got: %#v", got)
	}
	if gotStr != expectedValue {
		t.Errorf("Read %s from cache; expected %s", gotStr, expectedValue)
	}
}

func TestCacheReadEmpty(t *testing.T) {
	ch := New()

	got := ch.Read("key")
	if got != nil {
		t.Errorf("Expected nil payload when reading non-existent key")
	}
}

func TestCacheDelete(t *testing.T) {
	ch := New()
	key := "key"

	ch.Write(key, "value")
	ch.Delete(key)
	got := ch.Read(key)

	if got != nil {
		t.Errorf("Expected nil payload when reading deleted key")
	}
}
