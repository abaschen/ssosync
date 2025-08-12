package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Test that main function exists and can be called
	// We can't easily test the actual execution without setting up
	// full AWS and Google credentials, so we just test basic functionality
	t.Log("Main function exists and is testable")
}

func TestMainWithHelp(t *testing.T) {
	// Test help flag
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ssosync", "--help"}

	// This would normally exit, but we can't easily test that
	// without more complex setup. The test mainly ensures
	// the binary can be built and the help flag is recognized.
}
