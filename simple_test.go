package main

import "testing"

func TestPrint(t *testing.T) {
	Print(Wow{})
	if false {
		t.Errorf("Can't happen")
	}
}