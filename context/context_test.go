package context

import (
	"log/slog"
	"os"
	"testing"

	"github.com/alexandreh2ag/mib/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	cfg := &config.Config{}
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	fs := afero.NewMemMapFs()
	want := &Context{
		Config:     cfg,
		WorkingDir: "/app",
		Logger:     logger,
		LogLevel:   level,
		FS:         fs,
	}
	got := NewContext(cfg, "/app", logger, level, fs)

	assert.Equal(t, want, got)
}

func TestDefaultContext(t *testing.T) {
	workingDir, _ := os.Getwd()
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	cfg := config.DefaultConfig()
	want := &Context{
		Config:     &cfg,
		WorkingDir: workingDir,
		FS:         afero.NewOsFs(),
		Logger:     logger,
		LogLevel:   level,
	}
	got := DefaultContext()
	assert.Equal(t, want, got)
}
