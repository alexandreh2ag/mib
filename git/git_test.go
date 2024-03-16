package git

import (
	"errors"
	"fmt"
	billyAfero "github.com/Maldris/go-billy-afero"
	"github.com/alexandreh2ag/mib/context"
	mockgit "github.com/alexandreh2ag/mib/mock/git"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"os"
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

func TestCheckIfFileStaged_SuccessChangeDetected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Added},
		},
		nil,
	)
	got := CheckIfFileStaged(m)
	assert.True(t, got)
}

func TestCheckIfFileStaged_SuccessNoChangeDetected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(1).Return(
		git.Status{
			"foo/Dockerfile": &git.FileStatus{Worktree: git.Unmodified, Staging: git.Untracked},
		},
		nil,
	)
	got := CheckIfFileStaged(m)
	assert.False(t, got)
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

type FS struct {
	FS afero.Fs
}

type File struct {
	afero.File
}

func (f File) Lock() error {
	return nil
}

func (f File) Unlock() error {
	return nil
}

func (f FS) Create(filename string) (billy.File, error) {
	file, err := f.FS.Create(filename)
	if err != nil {
		return nil, err
	}
	bFile := File{File: file}
	return bFile, err
}

func (f FS) Open(filename string) (billy.File, error) {
	//TODO implement me
	panic("implement me")
}

func (f FS) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	//TODO implement me
	panic("implement me")
}

func (f FS) Stat(filename string) (os.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f FS) Rename(oldpath, newpath string) error {
	//TODO implement me
	panic("implement me")
}

func (f FS) Remove(filename string) error {
	//TODO implement me
	panic("implement me")
}

func (f FS) Join(elem ...string) string {
	//TODO implement me
	panic("implement me")
}

func (f FS) TempFile(dir, prefix string) (billy.File, error) {
	//TODO implement me
	panic("implement me")
}

func (f FS) ReadDir(path string) ([]os.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f FS) MkdirAll(filename string, perm os.FileMode) error {
	//TODO implement me
	panic("MkdirAll implement me")
}

func (f FS) Lstat(filename string) (os.FileInfo, error) {
	//TODO implement me
	panic("Lstat implement me")
}

func (f FS) Symlink(target, link string) error {
	//TODO implement me
	panic("Symlink implement me")
}

func (f FS) Readlink(link string) (string, error) {
	//TODO implement me
	panic("Readlink implement me")
}

func (f FS) Chroot(path string) (billy.Filesystem, error) {
	//TODO implement me
	panic("Chroot implement me")
}

func (f FS) Root() string {
	//TODO implement me
	panic("Root implement me")
}

func TestGit_CommitFileContent_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.WorkingDir = "/app/repo"
	want := "name: foo\ntag: 0.1"

	repo := initGitRepo(t, ctx)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/Dockerfile"), []byte("FROM debian:latest"), 0644)
	hash := stageAllAndCommit(t, repo, ctx.WorkingDir, "Add image")
	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	gitManger := &Git{r: repo, w: worktree}
	got, err := gitManger.CommitFileContent(&hash, "foo/mib.yml")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGit_CommitFileContent_HashNotExist(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.WorkingDir = "/app/repo"

	repo := initGitRepo(t, ctx)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/Dockerfile"), []byte("FROM debian:latest"), 0644)
	_ = stageAllAndCommit(t, repo, ctx.WorkingDir, "Add image")
	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	gitManger := &Git{r: repo, w: worktree}
	hash := plumbing.NewHash("wrong")
	_, err = gitManger.CommitFileContent(&hash, "foo/mib.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object not found")
}

func TestGit_CommitFileContent_PahtNotExistInCommit(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.WorkingDir = "/app/repo"

	repo := initGitRepo(t, ctx)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/Dockerfile"), []byte("FROM debian:latest"), 0644)
	hash := stageAllAndCommit(t, repo, ctx.WorkingDir, "Add image")
	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	gitManger := &Git{r: repo, w: worktree}
	_, err = gitManger.CommitFileContent(&hash, "foo/wrong.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestGit_AddWithOptions(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.WorkingDir = "/app/repo"

	repo := initGitRepo(t, ctx)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.1"), 0644)
	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	gitManger := &Git{r: repo, w: worktree}
	err = gitManger.AddWithOptions(&git.AddOptions{Path: "foo/mib.yml"})
	assert.NoError(t, err)
	status, err := worktree.Status()
	for path, fileStatus := range status {
		if path == "foo/mib.yml" {
			assert.Equal(t, git.Added, fileStatus.Staging)
		}
	}
}

func TestGit_GetCommitFilesChanged_SuccessLongHash(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.WorkingDir = "/app/repo"
	want := []string{"bar/Dockerfile", "bar/mib.yml", "foo/README.md", "foo/mib.yml", "foo/rootfs/file2"}

	repo := initGitRepo(t, ctx)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/Dockerfile"), []byte("FROM debian:latest"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/README.md"), []byte("# Readme"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/rootfs/file"), []byte("file"), 0644)
	_ = stageAllAndCommit(t, repo, ctx.WorkingDir, "Add image")

	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.2"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "bar/mib.yml"), []byte("name: bar\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "bar/Dockerfile"), []byte("FROM debian:latest"), 0644)
	_ = ctx.FS.Rename(filepath.Join(ctx.WorkingDir, "foo/rootfs/file"), filepath.Join(ctx.WorkingDir, "foo/rootfs/file2"))
	_ = ctx.FS.Remove(filepath.Join(ctx.WorkingDir, "foo/README.md"))
	hash := stageAllAndCommit(t, repo, ctx.WorkingDir, "Add image2")
	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	gitManger := &Git{r: repo, w: worktree}
	got, err := gitManger.GetCommitFilesChanged(hash.String())
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGit_GetCommitFilesChanged_SuccessShortHash(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.WorkingDir = "/app/repo"
	want := []string{"foo/Dockerfile", "foo/mib.yml"}

	repo := initGitRepo(t, ctx)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/Dockerfile"), []byte("FROM debian:latest"), 0644)
	hash := stageAllAndCommit(t, repo, ctx.WorkingDir, "Add image")

	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	gitManger := &Git{r: repo, w: worktree}
	got, err := gitManger.GetCommitFilesChanged(hash.String()[0:20])
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGit_GetCommitFilesChanged_ErrorGetCommit(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.WorkingDir = "/app/repo"

	repo := initGitRepo(t, ctx)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/mib.yml"), []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(ctx.FS, filepath.Join(ctx.WorkingDir, "foo/Dockerfile"), []byte("FROM debian:latest"), 0644)
	_ = stageAllAndCommit(t, repo, ctx.WorkingDir, "Add image")

	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	gitManger := &Git{r: repo, w: worktree}
	_, err = gitManger.GetCommitFilesChanged("wrong")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit not foun")
}

func initGitRepo(t *testing.T, ctx *context.Context) *git.Repository {
	memPath := filepath.Join(ctx.WorkingDir, git.GitDirName)
	memBaseDir := afero.NewBasePathFs(ctx.FS, memPath)
	fsBaseDir := afero.NewBasePathFs(ctx.FS, ctx.WorkingDir)
	mem := billyAfero.New(memBaseDir, memPath, false)
	fs := billyAfero.New(fsBaseDir, ctx.WorkingDir, false)
	_ = ctx.FS.Mkdir(filepath.Dir(ctx.WorkingDir), 0775)
	store := filesystem.NewStorage(mem, cache.NewObjectLRUDefault())
	repo, err := git.InitWithOptions(store, fs, git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName("main")})
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	if err != nil {
		panic(err)
	}
	repoConfig, _ := repo.Config()
	err = repo.Storer.SetConfig(repoConfig)
	assert.NoError(t, err)
	return repo
}

func stageAllAndCommit(t *testing.T, repo *git.Repository, path, message string) (hash plumbing.Hash) {
	worktree, err := repo.Worktree()
	assert.NoError(t, err)
	err = worktree.AddWithOptions(&git.AddOptions{All: true, Path: path})
	hash, err = worktree.Commit(message, &git.CommitOptions{})
	assert.NoError(t, err)
	return hash
}
