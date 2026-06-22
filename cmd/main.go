package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/bmf-san/gondola"
)

const defaultConfigPath = "config.yaml"

// resolveConfigPath determines the configuration file path. The GONDOLA_CONFIG
// environment variable takes precedence over the -config flag.
func resolveConfigPath(flagPath string) string {
	if env := os.Getenv("GONDOLA_CONFIG"); env != "" {
		return env
	}
	return flagPath
}

// openConfig opens the configuration file at the given path.
func openConfig(path string) (*os.File, error) {
	if path == "" {
		return nil, errors.New("config file is not specified")
	}
	// #nosec G304 -- path is operator-provided configuration, not user input.
	return os.Open(filepath.Clean(path))
}

// version returns the module version embedded by the build, or "dev".
func version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

// run parses arguments, builds Gondola and runs it, returning a process exit
// code.
func run(args []string) int {
	fs := flag.NewFlagSet("gondola", flag.ContinueOnError)
	cfgFlag := fs.String("config", defaultConfigPath, "path to configuration file")
	showVersion := fs.Bool("version", false, "print version information and exit")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if *showVersion {
		fmt.Printf("gondola %s\n", version())
		return 0
	}

	path := resolveConfigPath(*cfgFlag)
	f, err := openConfig(path)
	if err != nil {
		slog.Error("failed to open config file", slog.String("error", err.Error()))
		return 1
	}
	defer func() { _ = f.Close() }()

	opts := []gondola.Option{gondola.WithConfigPath(path)}
	if level := os.Getenv("GONDOLA_LOG_LEVEL"); level != "" {
		opts = append(opts, gondola.WithLogLevel(level))
	}

	g, err := gondola.NewGondola(f, opts...)
	if err != nil {
		slog.Error("failed to initialize gondola", slog.String("error", err.Error()))
		return 1
	}

	if err := g.Run(); err != nil {
		slog.Error("server stopped with error", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

func main() {
	defer func() {
		if x := recover(); x != nil {
			slog.Error("panic recovered", slog.String("stack", string(debug.Stack())))
			os.Exit(1)
		}
	}()
	os.Exit(run(os.Args[1:]))
}
