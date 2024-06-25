package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	os.Args = []string{"config", "../testdata/config/config.yml"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	err := parseFlags()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseFlagsError(t *testing.T) {
	os.Args = []string{"invalid"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	err := parseFlags()
	if err == nil {
		t.Error("expected error")
	}
}

func TestSetConfig(t *testing.T) {
	_, err := setConfig("")
	if err == nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = setConfig("invalid_file")
	if err == nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = setConfig("../testdata/config/config.yml")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
