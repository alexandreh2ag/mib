package context

import (
	"github.com/alexandreh2ag/mib/config"
	"github.com/spf13/afero"
	"log/slog"
	"os"
)

type Context struct {
	Config     *config.Config
	WorkingDir string
	Logger     *slog.Logger
	LogLevel   *slog.LevelVar
	FS         afero.Fs
}

func NewContext(config *config.Config, workingDir string, logger *slog.Logger, logLevel *slog.LevelVar, FSProvider afero.Fs) *Context {
	return &Context{Config: config, WorkingDir: workingDir, Logger: logger, LogLevel: logLevel, FS: FSProvider}
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
