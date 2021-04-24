package simple

import (
	"testing"
)

func TestPrint(t *testing.T) {
	Print(Wow{})
	if false {
		t.Errorf("Can't happen")
	}
}

func TestCheck(t *testing.T) {
	Check(Wow{})
	if false {
		t.Errorf("Can't happen")
	}
}
