package build

import (
	"errors"
	"github.com/alexandreh2ag/mib/container/docker"
	"github.com/alexandreh2ag/mib/context"
	mibGit "github.com/alexandreh2ag/mib/git"
	mockgit "github.com/alexandreh2ag/mib/mock/git"
	mock_types_container "github.com/alexandreh2ag/mib/mock/types/container"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"io"
	"testing"
)

func TestGetDirtyRunFn_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetDirtyCmd(ctx)
	cmd.Flags().Bool(PushImages, false, "")
	cmd.SetOut(io.Discard)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
		},
		nil,
	)
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
	builderDocker.EXPECT().BuildImages(gomock.Any()).Times(1).Return(nil)
	builderDocker.EXPECT().PushImages(gomock.Any()).Times(1).Return(nil)
	ctx.Builders[docker.KeyBuilder] = builderDocker
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)
	cmd.SetArgs([]string{
		"--" + PushImages,
	})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetDirtyRunFn_ErrorCreateGitManger(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetDirtyCmd(ctx)
	cmd.SetOut(io.Discard)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return nil, errors.New("error")
	}
	builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
	ctx.Builders[docker.KeyBuilder] = builderDocker
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestGetDirtyRunFn_ErrorBuildImages(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetDirtyCmd(ctx)
	cmd.SetOut(io.Discard)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
		},
		nil,
	)
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
	builderDocker.EXPECT().BuildImages(gomock.Any()).Times(1).Return(errors.New("error"))
	ctx.Builders[docker.KeyBuilder] = builderDocker
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestGetDirtyRunFn_SuccessPushImages(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetDirtyCmd(ctx)
	cmd.Flags().Bool(PushImages, false, "")
	cmd.SetOut(io.Discard)
	fsFake := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fsFake)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
		},
		nil,
	)
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
	builderDocker.EXPECT().BuildImages(gomock.Any()).Times(1).Return(nil)
	builderDocker.EXPECT().PushImages(gomock.Any()).Times(1).Return(errors.New("error"))
	ctx.Builders[docker.KeyBuilder] = builderDocker
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)
	cmd.SetArgs([]string{
		"--" + PushImages,
	})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}
