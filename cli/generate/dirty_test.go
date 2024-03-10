package generate

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"
	mibGit "github.com/alexandreh2ag/mib/git"
	mockgit "github.com/alexandreh2ag/mib/mock/git"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	worktree := mockgit.NewMockWorktree(ctrl)
	worktree.EXPECT().Status().Times(1).Return(
		git.Status{},
		nil,
	)
	mibGit.GetWorktree = func(ctx *context.Context) (mibGit.Worktree, error) {
		return worktree, nil
	}

	err := GetDirtyRunFn(ctx)(cmd, []string{})
	assert.NoError(t, err)
}

func TestGetCommitRunFn_ErrorWorkTree(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetDirtyCmd(ctx)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/app"
	_ = fsFake.Mkdir(path, 0775)
	mibGit.GetWorktree = func(ctx *context.Context) (mibGit.Worktree, error) {
		return nil, errors.New("error")
	}
	err := GetDirtyRunFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}
