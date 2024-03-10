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
