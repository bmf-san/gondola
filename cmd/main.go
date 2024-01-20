package main

import (
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

func main() {
	defer func() {
		if x := recover(); x != nil {
			slog.Error(string(debug.Stack()))
		}
		os.Exit(1)
	}()

	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	if cfgFile == "" {
		slog.Error("config file is not specified")
		os.Exit(1)
	}

	cfg, err := os.Open(filepath.Clean(cfgFile))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	defer func() {
		err = cfg.Close()
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	gondola, err := gondola.NewGondola(cfg)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	gondola.Run()
}
