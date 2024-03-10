package cli

import (
	"fmt"
	"github.com/alexandreh2ag/mib/config"
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_initConfig_SuccessConfigEmpty(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte(""), 0644)
	want := &config.Config{Build: config.Build{ExtensionExclude: ".md,.txt"}}
	initConfig(ctx, cmd)
	assert.Equal(t, want, ctx.Config)
}

func Test_initConfig_SuccessOverrideDefaultConfig(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("build: {extensionExclude: '.txt,.log'}\ntemplate: {imagePath: imageTmpl.tmpl, indexPath: indexTmpl.tmpl}"), 0644)
	want := &config.Config{
		Build: config.Build{
			ExtensionExclude: ".txt,.log",
		},
		Template: config.Template{
			ImagePath: "imageTmpl.tmpl",
			IndexPath: "indexTmpl.tmpl",
		},
	}
	initConfig(ctx, cmd)
	assert.Equal(t, want, ctx.Config)
}

func Test_initConfig_SuccessWithConfigFlag(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/foo"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/foo.yml", path), []byte("build: {extensionExclude: '.txt,.log'}"), 0644)
	want := &config.Config{
		Build: config.Build{
			ExtensionExclude: ".txt,.log",
		},
	}
	viper.Set(Config, fmt.Sprintf("%s/foo.yml", path))
	initConfig(ctx, cmd)
	assert.Equal(t, want, ctx.Config)
}

func Test_initConfig_FailWithUnmarshallError(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("build: []"), 0644)

	defer func() {
		if r := recover(); r != nil {
			assert.True(t, true)
		} else {
			t.Errorf("initConfig should have panicked")
		}
	}()
	initConfig(ctx, cmd)
}

func TestGetRootPreRunEFn_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	err := GetRootPreRunEFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "LevelVar(INFO)", ctx.LogLevel.String())
}

func TestGetRootPreRunEFn_SuccessWithWorkingDirFlag(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/foo"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("build: {}"), 0644)
	cmd.SetArgs([]string{
		"--" + WorkingDir, path},
	)
	_ = cmd.Execute()

	err := GetRootPreRunEFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "/foo", ctx.WorkingDir)
}

func TestGetRootPreRunEFn_SuccessWithLogLevelFlag(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("build: {}"), 0644)
	cmd.SetArgs([]string{
		"--" + LogLevel, "ERROR"},
	)
	_ = cmd.Execute()

	err := GetRootPreRunEFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "LevelVar(ERROR)", ctx.LogLevel.String())
}

func TestGetRootPreRunEFn_FailedWhenOverrideTmpl(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("template: {indexPath: 'wrong'}"), 0644)
	_ = cmd.Execute()

	err := GetRootPreRunEFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open /app/wrong: file does not exist")
}

func TestGetRootPreRunEFn_FailedWithLogLevelFlag(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetRootCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("build: {}"), 0644)
	cmd.SetArgs([]string{
		"--" + LogLevel, "WRONG"},
	)
	_ = cmd.Execute()

	err := GetRootPreRunEFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slog: level string \"WRONG\": unknown name")
}
