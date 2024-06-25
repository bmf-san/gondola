package main

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/bmf-san/gondola"
)

var cfgFile string

func init() {
	flag.StringVar(&cfgFile, "config", "config.yaml", "config file path")

}

// parseFlags parses the command line flags.
func parseFlags() error {
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return errors.New("config file is not specified")
	}
	return nil
}

// setConfig returns the config file.
func setConfig(cfgFile string) (*os.File, error) {
	if cfgFile == "" {
		return nil, errors.New("config file is not specified")
	}

	cfg, err := os.Open(filepath.Clean(cfgFile))
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func main() {
	defer func() {
		if x := recover(); x != nil {
			slog.Error(string(debug.Stack()))
		}
		os.Exit(1)
	}()

	err := parseFlags()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	cfg, err := setConfig(cfgFile)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	gondola, err := gondola.NewGondola(cfg)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = gondola.Run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
