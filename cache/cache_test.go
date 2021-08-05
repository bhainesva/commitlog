package cache

import (
	"testing"
)

func TestRead(t *testing.T) {
	expectedValue := "test"
	key := "key"
	ch := From(map[string]interface{}{key: expectedValue})

	got := ch.Read(key)
	gotStr, ok := got.(string)
	if !ok {
		t.Errorf("Expected to find string value on cache read, got: %#v", got)
	}
	if gotStr != expectedValue {
		t.Errorf("Read %s from cache; expected %s", gotStr, expectedValue)
	}
}

func TestWrite(t *testing.T) {
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
	key := "key"
	ch := From(map[string]interface{}{key: "value"})

	ch.Delete(key)
	got := ch.Read(key)

	if got != nil {
		t.Errorf("Expected nil payload when reading deleted key")
	}
}
