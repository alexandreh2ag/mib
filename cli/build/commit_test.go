package build

import (
	"errors"
	"github.com/alexandreh2ag/mib/container/docker"
	"github.com/alexandreh2ag/mib/context"
	mibGit "github.com/alexandreh2ag/mib/git"
	mockgit "github.com/alexandreh2ag/mib/mock/git"
	mock_types_container "github.com/alexandreh2ag/mib/mock/types/container"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"io"
	"testing"
)

func TestGetCommitRunFn(t *testing.T) {

	tests := []struct {
		name      string
		cmdArgs   []string
		imageData string
		preFn     func(ctx *context.Context, ctrl *gomock.Controller)
		checkFn   func(t *testing.T, err error)
	}{
		{
			name:      "Success",
			imageData: "name: foo\ntag: 0.1",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				m := mockgit.NewMockManager(ctrl)
				m.EXPECT().GetCommitFilesChanged(gomock.Eq("xxx")).Times(1).Return(
					[]string{"foo/Dockerfile"},
					nil,
				)
				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return m, nil
				}

				builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
				builderDocker.EXPECT().BuildImages(gomock.Any(), gomock.Eq(true)).Times(1).Return(nil)
				ctx.Builders[docker.KeyBuilder] = builderDocker
			},
			cmdArgs: []string{
				"--" + PushImages,
				"--" + Commit, "xxx",
			},
			checkFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:      "SuccessWithoutCommitFlag",
			imageData: "name: foo\ntag: 0.1",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				m := mockgit.NewMockManager(ctrl)

				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return m, nil
				}
				gomock.InOrder(
					m.EXPECT().Head().Times(1).Return("xxx", nil),
					m.EXPECT().GetCommitFilesChanged(gomock.Eq("xxx")).Times(1).Return(
						[]string{"foo/Dockerfile"},
						nil,
					),
				)

				builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
				builderDocker.EXPECT().BuildImages(gomock.Any(), gomock.Eq(true)).Times(1).Return(nil)
				ctx.Builders[docker.KeyBuilder] = builderDocker
			},
			cmdArgs: []string{
				"--" + PushImages,
			},
			checkFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:      "ErrorCreateGitManger",
			imageData: "name: foo\ntag: 0.1",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return nil, errors.New("error")
				}

				builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
				ctx.Builders[docker.KeyBuilder] = builderDocker
			},
			cmdArgs: []string{"--" + Commit, "xxx"},
			checkFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error")
			},
		},
		{
			name:      "ErrorGetHeadReference",
			imageData: "name: foo\ntag: 0.1",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				m := mockgit.NewMockManager(ctrl)
				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return m, nil
				}
				m.EXPECT().Head().Times(1).Return("", errors.New("error"))
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
			name:      "FailLoadImages",
			imageData: "name: foo\ntag: ",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				m := mockgit.NewMockManager(ctrl)
				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return m, nil
				}

				builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
				ctx.Builders[docker.KeyBuilder] = builderDocker
			},
			cmdArgs: []string{"--" + Commit, "xxx"},
			checkFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "configuration file is not valid")
			},
		},
		{
			name:      "ErrorGetChangedFiles",
			imageData: "name: foo\ntag: 0.1",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				m := mockgit.NewMockManager(ctrl)
				m.EXPECT().GetCommitFilesChanged(gomock.Eq("xxx")).Times(1).Return(
					nil,
					errors.New("error"),
				)
				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return m, nil
				}

				builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
				ctx.Builders[docker.KeyBuilder] = builderDocker
			},
			cmdArgs: []string{"--" + Commit, "xxx"},
			checkFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error")
			},
		},
		{
			name:      "ErrorBuildImages",
			imageData: "name: foo\ntag: 0.1",
			preFn: func(ctx *context.Context, ctrl *gomock.Controller) {
				m := mockgit.NewMockManager(ctrl)
				m.EXPECT().GetCommitFilesChanged(gomock.Eq("xxx")).Times(1).Return(
					[]string{"foo/Dockerfile"},
					nil,
				)
				mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
					return m, nil
				}

				builderDocker := mock_types_container.NewMockBuilderImage(ctrl)
				builderDocker.EXPECT().BuildImages(gomock.Any(), gomock.Eq(false)).Times(1).Return(errors.New("error"))
				ctx.Builders[docker.KeyBuilder] = builderDocker
			},
			cmdArgs: []string{"--" + Commit, "xxx"},
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
			cmd := GetCommitCmd(ctx)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.Flags().Bool(PushImages, false, "")
			viper.Reset()
			viper.SetFs(ctx.FS)

			_ = ctx.FS.Mkdir(ctx.WorkingDir, 0775)
			_ = afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte(tt.imageData), 0644)
			_ = afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)

			tt.preFn(ctx, ctrl)

			cmd.SetArgs(tt.cmdArgs)
			err := cmd.Execute()
			tt.checkFn(t, err)
		})
	}
}
