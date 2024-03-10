package git

import (
	"errors"
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	mockgit "github.com/alexandreh2ag/mib/mock/git"
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"path/filepath"
	"testing"
)

func TestGetStageFilesChanged_Success(t *testing.T) {
	ctx := context.DefaultContext()
	workgingDir, _ := filepath.Abs(fmt.Sprintf("%s/..", ctx.WorkingDir))
	ctx.WorkingDir = workgingDir
	worktree, err := GetWorktree(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, worktree)
}

func TestGetStageFilesChanged_Fail(t *testing.T) {
	ctx := context.DefaultContext()
	worktree, err := GetWorktree(ctx)
	assert.Error(t, err)
	assert.Nil(t, worktree)
}

func TestGetStageFilesChanged(t *testing.T) {
	ctx := context.TestContext(nil)

	tests := []struct {
		name        string
		getWorktree func(ctrl *gomock.Controller) func(ctx *context.Context) (Worktree, error)
		want        []string
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			getWorktree: func(ctrl *gomock.Controller) func(ctx *context.Context) (Worktree, error) {
				worktree := mockgit.NewMockWorktree(ctrl)
				worktree.EXPECT().Status().Times(1).Return(
					git.Status{
						"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
						"foo/unmodified": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Unmodified},
						"foo/remove":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Renamed, Extra: "foo/remove"},
					},
					nil,
				)
				return func(ctx *context.Context) (Worktree, error) {
					return worktree, nil
				}
			},
			want:    []string{"foo/Dockerfile", "foo/remove"},
			wantErr: assert.NoError,
		},
		{
			name: "Fail",
			getWorktree: func(ctrl *gomock.Controller) func(ctx *context.Context) (Worktree, error) {
				return func(ctx *context.Context) (Worktree, error) {
					return nil, errors.New("error")
				}
			},
			want:    []string{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			GetWorktree = tt.getWorktree(ctrl)
			got, err := GetStageFilesChanged(ctx)
			tt.wantErr(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
