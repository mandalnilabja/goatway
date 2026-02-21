package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/version"
)

func setupLogger() *slog.Logger {
	// Use sensible defaults: info level, text format
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

func printStartupBanner(cfg *config.Config) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🐐 Goatway %s - Local OpenAI-Compatible Proxy\n", version.Version)
	fmt.Fprintln(os.Stderr, "════════════════════════════════════════════════")
	if cfg.EnableWebUI {
		fmt.Fprintf(os.Stderr, "Web UI:     http://localhost%s/web\n", cfg.ServerPort)
	}
	fmt.Fprintf(os.Stderr, "Proxy API:  http://localhost%s/v1/chat/completions\n", cfg.ServerPort)
	fmt.Fprintf(os.Stderr, "Admin API:  http://localhost%s/api/admin/\n", cfg.ServerPort)
	fmt.Fprintf(os.Stderr, "Data:       %s\n", config.DataDir())
	fmt.Fprintln(os.Stderr, "════════════════════════════════════════════════")
	fmt.Fprintf(os.Stderr, "\n")
}
