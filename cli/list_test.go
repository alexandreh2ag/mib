package cli

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetListRunFn_Success_NoImage(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetListCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)

	err := GetListRunFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
}

func TestGetListRunFn_Success_WithImages(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetListCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/bar/mib.yml", []byte("name: bar\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/bar/Dockerfile", []byte("FROM debian:latest"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/foo-bar/mib.yml", []byte("name: foo-bar\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/foo-bar/Dockerfile", []byte("FROM foo:0.1"), 0644)

	err := GetListRunFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
}
