package generate

import (
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetIndexRunFn_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetIndexCmd(ctx)
	afs := &afero.Afero{Fs: ctx.FS}
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)

	err := GetIndexRunFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
	content, err := afs.ReadFile(fmt.Sprintf("%s/README.md", ctx.WorkingDir))
	assert.NoError(t, err)
	assert.Contains(t, string(content), "# Images list")

}

func TestGetIndexRunFn_FailLoadImage(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetIndexCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/foo/mib.yml", path), []byte("name: foo\ntag: "), 0644)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/foo/Dockerfile", path), []byte("FROM debian:latest"), 0644)

	err := GetIndexRunFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "images configuration file is not valid")

}
