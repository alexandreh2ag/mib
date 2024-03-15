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

func TestGetDirtyRunFn(t *testing.T) {

	tests := []struct {
		name    string
		cmdArgs []string
		preFn   func(ctx *context.Context, ctrl *gomock.Controller)
		checkFn func(t *testing.T, err error)
	}{
		{
			name: "Success",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
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
			},
			cmdArgs: []string{"--" + PushImages},
			checkFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "ErrorCreateGitManger",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return nil, errors.New("error")
				}

				builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
				ctx.Builders[docker.KeyBuilder] = builderDocker
			},
			cmdArgs: []string{},
			checkFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error")
			},
		},
		{
			name: "ErrorBuildImages",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
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
			},
			cmdArgs: []string{},
			checkFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error")
			},
		},
		{
			name: "ErrorPushImages",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
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
			},
			cmdArgs: []string{"--" + PushImages},
			checkFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TestContext(nil)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cmd := GetDirtyCmd(ctx)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.Flags().Bool(PushImages, false, "")
			viper.Reset()
			viper.SetFs(ctx.FS)

			_ = ctx.FS.Mkdir(ctx.WorkingDir, 0775)
			_ = afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
			_ = afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)

			tt.preFn(ctx, ctrl)

			cmd.SetArgs(tt.cmdArgs)
			err := cmd.Execute()
			tt.checkFn(t, err)
		})
	}
}
