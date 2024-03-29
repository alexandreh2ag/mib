package context

import (
	"github.com/alexandreh2ag/mib/config"
	typesContainers "github.com/alexandreh2ag/mib/types/container"
	"github.com/spf13/afero"
	"io"
	"log/slog"
	"os"
)

type Context struct {
	Config     *config.Config
	WorkingDir string
	Logger     *slog.Logger
	LogLevel   *slog.LevelVar
	FS         afero.Fs
	Builders   typesContainers.Builders
}

func NewContext(config *config.Config, workingDir string, logger *slog.Logger, logLevel *slog.LevelVar, FSProvider afero.Fs) *Context {
	return &Context{Config: config, WorkingDir: workingDir, Logger: logger, LogLevel: logLevel, FS: FSProvider, Builders: typesContainers.Builders{}}
}

func DefaultContext() *Context {
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cfg := config.DefaultConfig()
	return NewContext(&cfg, workingDir, slog.New(slog.NewTextHandler(os.Stdout, opts)), level, afero.NewOsFs())
}

func TestContext(logBuffer io.Writer) *Context {
	if logBuffer == nil {
		logBuffer = io.Discard
	}
	cfg := config.DefaultConfig()
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}

	return &Context{
		Logger:     slog.New(slog.NewTextHandler(logBuffer, opts)),
		LogLevel:   level,
		Config:     &cfg,
		FS:         afero.NewMemMapFs(),
		WorkingDir: "/app",
		Builders:   typesContainers.Builders{},
	}
}
