package ui

import (
	"testing"
)

func TestProgressBar_Add(t *testing.T) {
	size := int64(100)
	label := "Test Progress"
	pb := NewProgressBar(size, label)

	// Add some increments within range; should not error
	if err := pb.Add(10); err != nil {
		t.Errorf("Add(10) returned unexpected error: %v", err)
	}

	if err := pb.Add(0); err != nil {
		t.Errorf("Add(0) returned unexpected error: %v", err)
	}

	// Add more increments, up to the size; no errors expected
	if err := pb.Add(90); err != nil {
		t.Errorf("Add(90) returned unexpected error: %v", err)
	}

	// Add beyond max size; library saturates and returns no error
	if err := pb.Add(10); err != nil {
		t.Errorf("Add(10) beyond max size returned unexpected error: %v", err)
	}
}
