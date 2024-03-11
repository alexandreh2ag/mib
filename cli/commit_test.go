package cli

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"
	mibGit "github.com/alexandreh2ag/mib/git"
	mockgit "github.com/alexandreh2ag/mib/mock/git"
	"github.com/alexandreh2ag/mib/template"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func initFS(fs afero.Fs) {
	_ = afero.WriteFile(fs, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
	_ = afero.WriteFile(fs, "/app/foo/Dockerfile", []byte("FROM debian:latest"), 0644)
	_ = afero.WriteFile(fs, "/app/bar/mib.yml", []byte("name: bar\ntag: 0.1"), 0644)
	_ = afero.WriteFile(fs, "/app/bar/Dockerfile", []byte("FROM debian:latest"), 0644)
	_ = afero.WriteFile(fs, "/app/foo-bar/mib.yml", []byte("name: foo-bar\ntag: 0.1"), 0644)
	_ = afero.WriteFile(fs, "/app/foo-bar/Dockerfile", []byte("FROM foo:0.1"), 0644)
}

func TestGetCommitRunFn_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(4).Return(
		git.Status{
			"foo/Dockerfile":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"bar/test.yml":       &git.FileStatus{Staging: git.Added},
			"foo-bar/mib.yml":    &git.FileStatus{Staging: git.Added},
			"foo-bar/Dockerfile": &git.FileStatus{Staging: git.Added},
			"foo-bar/file":       &git.FileStatus{Worktree: git.Untracked},
			"bar-bar/mib.yml":    &git.FileStatus{Staging: git.Deleted},
		},
		nil,
	)
	hash := plumbing.NewHash("hash")
	m.EXPECT().ResolveRevision(gomock.Any()).Times(1).Return(&hash, nil)
	m.EXPECT().CommitFileContent(gomock.Any(), gomock.Eq("bar-bar/mib.yml")).Times(1).Return("name: bar-bar\ntag: 0.1", nil)

	gomock.InOrder(
		m.EXPECT().AddWithOptions(gomock.Eq(&git.AddOptions{Path: "README.md"})).Times(1).Return(nil),
		m.EXPECT().AddWithOptions(gomock.Any()).Times(3).Return(nil),
	)
	message := "This commit add 1, update 2 and 1 removed images\n\n+ foo-bar:0.1\n* bar:0.1\n* foo:0.1\n- bar-bar:0.1\n"
	m.EXPECT().CreateCommit(gomock.Eq(message), gomock.Any()).Times(1).Return(plumbing.NewHash("xxx"), nil)

	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	cmd.SetArgs([]string{
		"--" + generateIndex,
		"--" + generateImage,
	})
	err := cmd.Execute()
	assert.NoError(t, err)
	for _, pathFile := range []string{"/app/README.md", "/app/foo/README.md", "/app/bar/README.md", "/app/foo-bar/README.md"} {
		exist, errExist := afero.Exists(ctx.FS, pathFile)
		assert.NoErrorf(t, errExist, "got error when check %s exist: %v", pathFile, errExist)
		assert.Truef(t, exist, "file %s must exist", pathFile)
	}
}

func TestGetCommitRunFn_ErrorGetRepository(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)

	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return nil, errors.New("error")
	}

	err := GetCommitRunFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestGetCommitRunFn_ErrorGetImagesAdded(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	_ = afero.WriteFile(ctx.FS, "/app/foo-bar/mib.yml", []byte("{]"), 0644)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(3).Return(
		git.Status{
			"foo/Dockerfile":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"bar/test.yml":       &git.FileStatus{Staging: git.Added},
			"foo-bar/mib.yml":    &git.FileStatus{Staging: git.Added},
			"foo-bar/Dockerfile": &git.FileStatus{Staging: git.Added},
			"bar-bar/mib.yml":    &git.FileStatus{Staging: git.Deleted},
		},
		nil,
	)

	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}

	err := GetCommitRunFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse foo-bar/mib.yml with error")
}

func TestGetCommitRunFn_ErrorGetImagesRemoved(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(4).Return(
		git.Status{
			"foo/Dockerfile":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"bar/test.yml":       &git.FileStatus{Staging: git.Added},
			"foo-bar/mib.yml":    &git.FileStatus{Staging: git.Added},
			"foo-bar/Dockerfile": &git.FileStatus{Staging: git.Added},
			"bar-bar/mib.yml":    &git.FileStatus{Staging: git.Deleted},
		},
		nil,
	)
	hash := plumbing.NewHash("hash")
	m.EXPECT().ResolveRevision(gomock.Any()).Times(1).Return(&hash, nil)
	m.EXPECT().CommitFileContent(gomock.Any(), gomock.Eq("bar-bar/mib.yml")).Times(1).Return("{]", nil)
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}

	err := GetCommitRunFn(ctx)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse removed bar-bar/mib.yml with error")
}

func TestGetCommitRunFn_ErrorAddToStage(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(5).Return(
		git.Status{
			"foo/Dockerfile":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"bar/test.yml":       &git.FileStatus{Worktree: git.Modified},
			"foo-bar/mib.yml":    &git.FileStatus{Staging: git.Added},
			"foo-bar/Dockerfile": &git.FileStatus{Staging: git.Added},
		},
		nil,
	)

	m.EXPECT().AddWithOptions(gomock.Any()).Times(1).Return(errors.New("error"))
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	cmd.SetArgs([]string{
		"--" + gitStageAll,
	})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "add file bar/test.yml fail with error")
}

func TestGetCommitRunFn_ErrorGenerateIndex(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(4).Return(
		git.Status{
			"foo/Dockerfile":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"bar/test.yml":       &git.FileStatus{Worktree: git.Modified},
			"foo-bar/mib.yml":    &git.FileStatus{Staging: git.Added},
			"foo-bar/Dockerfile": &git.FileStatus{Staging: git.Added},
		},
		nil,
	)
	template.IndexTmplPath = "wrong.tmpl"
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	cmd.SetArgs([]string{
		"--" + generateIndex,
	})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template wrong.tmpl not found")
}

func TestGetCommitRunFn_ErrorGenerateImage(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(4).Return(
		git.Status{
			"foo/Dockerfile":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"bar/test.yml":       &git.FileStatus{Worktree: git.Modified},
			"foo-bar/mib.yml":    &git.FileStatus{Staging: git.Added},
			"foo-bar/Dockerfile": &git.FileStatus{Staging: git.Added},
		},
		nil,
	)
	template.ImageTmplPath = "wrong.tmpl"
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	cmd.SetArgs([]string{
		"--" + generateImage,
	})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template wrong.tmpl not found")
}

func TestGetCommitRunFn_ErrorGenerateCommitMessage(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(4).Return(
		git.Status{},
		nil,
	)
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no change detected to generate commit message")
}

func TestGetCommitRunFn_ErrorCreateCommit(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetCommitCmd(ctx)
	viper.Reset()
	viper.SetFs(ctx.FS)
	path := ctx.WorkingDir
	_ = ctx.FS.Mkdir(path, 0775)
	initFS(ctx.FS)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mockgit.NewMockManager(ctrl)
	m.EXPECT().Status().Times(4).Return(
		git.Status{
			"foo/Dockerfile":     &git.FileStatus{Worktree: git.Unmodified, Staging: git.Modified},
			"bar/test.yml":       &git.FileStatus{Worktree: git.Modified},
			"foo-bar/mib.yml":    &git.FileStatus{Staging: git.Added},
			"foo-bar/Dockerfile": &git.FileStatus{Staging: git.Added},
		},
		nil,
	)
	m.EXPECT().CreateCommit(gomock.Any(), gomock.Any()).Times(1).Return(plumbing.ZeroHash, errors.New("error"))
	mibGit.CreateGit = func(ctx *context.Context) (mibGit.Manager, error) {
		return m, nil
	}
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestSliceDifference(t *testing.T) {
	type args struct {
		s1 []string
		s2 []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "SuccessEmpty",
			args: args{
				s1: []string{},
				s2: []string{},
			},
			want: []string{},
		},
		{
			name: "SuccessCase1",
			args: args{
				s1: []string{"foo", "bar"},
				s2: []string{"bar"},
			},
			want: []string{"foo"},
		},
		{
			name: "SuccessCase2",
			args: args{
				s1: []string{"foo"},
				s2: []string{"bar"},
			},
			want: []string{"foo"},
		},
		{
			name: "SuccessCase3",
			args: args{
				s1: []string{},
				s2: []string{"bar"},
			},
			want: []string{},
		},
		{
			name: "SuccessCase4",
			args: args{
				s1: []string{"foo"},
				s2: []string{},
			},
			want: []string{"foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sliceDifference(tt.args.s1, tt.args.s2), "sliceDifference(%v, %v)", tt.args.s1, tt.args.s2)
		})
	}
}

func Test_sliceAddPrefix(t *testing.T) {
	type args struct {
		s      []string
		prefix string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "SuccessNothing",
			args: args{
				s:      []string{},
				prefix: "+",
			},
			want: []string{},
		},
		{
			name: "Success",
			args: args{
				s:      []string{"foo"},
				prefix: "+",
			},
			want: []string{"+foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sliceAddPrefix(tt.args.s, tt.args.prefix), "sliceAddPrefix(%v, %v)", tt.args.s, tt.args.prefix)
		})
	}
}

func Test_getCommitMessage_Success(t *testing.T) {
	nameImagesAdded := []string{"foo"}
	nameImagesUpdated := []string{"bar"}
	nameImagesRemoved := []string{"foo-old"}
	want := "This commit add 1, update 1 and 1 removed images\n\n+ foo\n* bar\n- foo-old\n"

	got, err := getCommitMessage(nameImagesAdded, nameImagesUpdated, nameImagesRemoved)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_getCommitMessage_SuccessMultiple(t *testing.T) {
	nameImagesAdded := []string{"foo", "fooo"}
	nameImagesUpdated := []string{"bar", "barr"}
	nameImagesRemoved := []string{"foo-old"}
	want := "This commit add 2, update 2 and 1 removed images\n\n+ foo\n+ fooo\n* bar\n* barr\n- foo-old\n"

	got, err := getCommitMessage(nameImagesAdded, nameImagesUpdated, nameImagesRemoved)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_getCommitMessage_Fail(t *testing.T) {
	nameImagesAdded := []string{}
	nameImagesUpdated := []string{}
	nameImagesRemoved := []string{}
	want := ""

	got, err := getCommitMessage(nameImagesAdded, nameImagesUpdated, nameImagesRemoved)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no change detected to generate commit message")
	assert.Equal(t, want, got)
}
