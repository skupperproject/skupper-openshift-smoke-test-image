package main

import "testing"

func TestSkupperOCPSmoke(t *testing.T) {
	err := run()
	if err != nil {
		t.Errorf("Test failed - %s", err.Error())
	}
}
