package generate

import (
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/template"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAllRunFn_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetAllCmd(ctx)
	fsFake := ctx.FS
	viper.Reset()
	viper.SetFs(fsFake)
	afs := &afero.Afero{Fs: ctx.FS}
	path := ctx.WorkingDir
	_ = fsFake.Mkdir(path, 0775)

	err := GetAllRunFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
	exist, err := afs.Exists(fmt.Sprintf("%s/README.md", path))
	assert.NoError(t, err)
	assert.True(t, exist)
}

func TestGetAllRunFn_ErrorWIthImagesTemplates(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetAllCmd(ctx)
	fsFake := ctx.FS
	viper.Reset()
	viper.SetFs(fsFake)
	afs := &afero.Afero{Fs: ctx.FS}
	path := ctx.WorkingDir
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/foo/mib.yml", path), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/foo/Dockerfile", path), []byte("FROM debian:latest"), 0644)
	template.ImageTmplPath = "no-exist.tmpl"
	err := GetAllRunFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template no-exist.tmpl not found")
	exist, err := afs.Exists(fmt.Sprintf("%s/README.md", path))
	assert.NoError(t, err)
	assert.False(t, exist)
}
