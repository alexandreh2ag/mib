package generate

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCommitRunFn_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetDirtyCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)

	err := GetDirtyRunFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "implement me")
}
