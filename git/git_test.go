package git

import (
	"errors"
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	mockgit "github.com/alexandreh2ag/mib/mock/git"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"path/filepath"
	"testing"
)

func TestCreateGit_Success(t *testing.T) {
	ctx := context.DefaultContext()
	workgingDir, _ := filepath.Abs(fmt.Sprintf("%s/..", ctx.WorkingDir))
	ctx.WorkingDir = workgingDir
	got, err := CreateGit(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateGit_ErrorCreateRpository(t *testing.T) {
	ctx := context.TestContext(nil)
	got, err := CreateGit(ctx)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestGit_Status(t *testing.T) {
	want := git.Status{
		"foo/mib.yml": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	w := mockgit.NewMockWorktree(ctrl)
	w.EXPECT().Status().Times(1).Return(want, nil)

	g := Git{w: w}
	got, err := g.Status()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGit_ResolveRevision(t *testing.T) {

	hash := plumbing.NewHash("hash")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mockgit.NewMockRepository(ctrl)
	r.EXPECT().ResolveRevision(gomock.Any()).Times(1).Return(&hash, nil)

	g := Git{r: r}
	got, err := g.ResolveRevision("HEAD^:foo")
	assert.NoError(t, err)
	assert.Equal(t, &hash, got)
}

func TestGit_CreateCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	w := mockgit.NewMockWorktree(ctrl)
	hash := plumbing.NewHash("xxx")
	w.EXPECT().Commit(gomock.Any(), gomock.Any()).Times(1).Return(hash, nil)

	g := Git{w: w}
	got, err := g.CreateCommit("test", &git.CommitOptions{All: false, AllowEmptyCommits: false})
	assert.NoError(t, err)
	assert.Equal(t, hash, got)
}

func TestGetStageFilesChanged_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	want := []string{"foo/Dockerfile", "foo/remove"}
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
			"foo/unmodified": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Unmodified},
			"foo/remove":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Renamed, Extra: "foo/remove"},
		},
		nil,
	)
	got := GetStageFilesChanged(m)
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameAdded_SuccessNothing(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"foo/test.yml":   &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
		},
		nil,
	)
	want := []string{}
	got, err := GetImagesNameAdded(ctx, m)
	assert.NoError(t, err)
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameAdded_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/foo/mib.yml", ctx.WorkingDir), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/bar/mib.yml", ctx.WorkingDir), []byte("name: bar\ntag: 0.1"), 0644)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
			"bar/mib.yml": &git.FileStatus{Worktree: git.Untracked, Staging: git.Untracked},
		},
		nil,
	)

	want := []string{"foo:0.1", "bar:0.1"}
	got, err := GetImagesNameAdded(ctx, m)
	assert.NoError(t, err)
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameAdded_ErrorFindImageFile(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
		},
		nil,
	)

	want := []string{}
	got, err := GetImagesNameAdded(ctx, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not load file foo/mib.yml")
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameAdded_ErrorMarshalImage(t *testing.T) {
	ctx := context.TestContext(nil)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/foo/mib.yml", ctx.WorkingDir), []byte("[}"), 0644)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
		},
		nil,
	)

	want := []string{}
	got, err := GetImagesNameAdded(ctx, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse foo/mib.yml with error")
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameRemoved_SuccessNothing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	want := []string{}
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
		},
		nil,
	)

	got, err := GetImagesNameRemoved(m)
	assert.NoError(t, err)
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameRemoved_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	want := []string{"foo:0.1"}

	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Staging: git.Deleted},
		},
		nil,
	)
	hash := plumbing.NewHash("hash")
	m.EXPECT().ResolveRevision(gomock.Any()).Times(1).Return(&hash, nil)
	m.EXPECT().CommitFileContent(gomock.Any(), gomock.Eq("foo/mib.yml")).Times(1).Return("name: foo\ntag: 0.1", nil)

	got, err := GetImagesNameRemoved(m)
	assert.NoError(t, err)
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameRemoved_ErrorRevision(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	want := []string{}

	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Staging: git.Deleted},
		},
		nil,
	)
	m.EXPECT().ResolveRevision(gomock.Any()).Times(1).Return(nil, errors.New("error"))

	got, err := GetImagesNameRemoved(m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameRemoved_ErrorContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	want := []string{}

	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Staging: git.Deleted},
		},
		nil,
	)
	hash := plumbing.NewHash("hash")
	m.EXPECT().ResolveRevision(gomock.Any()).Times(1).Return(&hash, nil)
	m.EXPECT().CommitFileContent(gomock.Any(), gomock.Eq("foo/mib.yml")).Times(1).Return("", errors.New("error"))

	got, err := GetImagesNameRemoved(m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
	assert.ElementsMatch(t, want, got)
}

func TestGetImagesNameRemoved_ErrorMarshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	want := []string{}

	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Staging: git.Deleted},
		},
		nil,
	)
	hash := plumbing.NewHash("hash")
	m.EXPECT().ResolveRevision(gomock.Any()).Times(1).Return(&hash, nil)
	m.EXPECT().CommitFileContent(gomock.Any(), gomock.Eq("foo/mib.yml")).Times(1).Return("[}", nil)

	got, err := GetImagesNameRemoved(m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse removed foo/mib.yml with error")
	assert.ElementsMatch(t, want, got)
}

func TestHasUntrackedFiles_SuccessNothing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified},
		},
		nil,
	)
	got := HasUntrackedFiles(m)
	assert.False(t, got)
}

func TestHasUntrackedFiles_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Untracked},
		},
		nil,
	)
	got := HasUntrackedFiles(m)
	assert.True(t, got)
}

func TestAddModifiedFilesToStage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/file.yml": &git.FileStatus{Worktree: git.Added},
			"foo/mib.yml":  &git.FileStatus{Worktree: git.Modified},
		},
		nil,
	)
	m.EXPECT().AddWithOptions(gomock.Any()).Times(1).Return(nil)
	err := AddModifiedFilesToStage(m)
	assert.NoError(t, err)
}

func TestAddModifiedFilesToStage_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/mib.yml": &git.FileStatus{Worktree: git.Modified},
		},
		nil,
	)
	m.EXPECT().AddWithOptions(gomock.Any()).Times(1).Return(errors.New("error"))
	err := AddModifiedFilesToStage(m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestAddFileToStage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)

	m.EXPECT().AddWithOptions(gomock.Any()).Times(1).Return(nil)
	err := AddFileToStage(m, "foo.txt")
	assert.NoError(t, err)
}
