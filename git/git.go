package git

import (
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/loader"
	"github.com/alexandreh2ag/mib/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

const (
	Added    = "+"
	Updated  = "*"
	Removed  = "-"
	NoUpdate = ""
)

type Repository interface {
	Worktree() (*git.Worktree, error)
	ResolveRevision(in plumbing.Revision) (*plumbing.Hash, error)
	CommitObject(h plumbing.Hash) (*object.Commit, error)
	CommitObjects() (object.CommitIter, error)
	Head() (*plumbing.Reference, error)
}

type Worktree interface {
	Status() (git.Status, error)
	AddWithOptions(opts *git.AddOptions) error
	Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
}

type Manager interface {
	Head() (string, error)
	Status() (git.Status, error)
	ResolveRevision(in plumbing.Revision) (*plumbing.Hash, error)
	CommitFileContent(hash *plumbing.Hash, path string) (string, error)
	AddWithOptions(opts *git.AddOptions) error
	CreateCommit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
	GetCommitFilesChanged(hash string) ([]string, error)
}

var _ Manager = &Git{}

type Git struct {
	r Repository
	w Worktree
}

func (g Git) Head() (string, error) {
	head, err := g.r.Head()
	if err != nil {
		return "", err
	}
	return head.String(), nil
}

func (g Git) CreateCommit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	return g.w.Commit(msg, opts)
}

func (g Git) Status() (git.Status, error) {
	return g.w.Status()
}

func (g Git) ResolveRevision(in plumbing.Revision) (*plumbing.Hash, error) {
	return g.r.ResolveRevision(in)
}

func (g Git) CommitFileContent(hash *plumbing.Hash, path string) (string, error) {
	commitObject, err := g.r.CommitObject(*hash)
	if err != nil {
		return "", err
	}

	file, err := commitObject.File(path)
	if err != nil {
		return "", err
	}
	return file.Contents()
}

func (g Git) AddWithOptions(opts *git.AddOptions) error {
	return g.w.AddWithOptions(opts)
}

func (g Git) GetCommitFilesChanged(hash string) ([]string, error) {
	var commit *object.Commit
	var err error
	files := []string{}
	if len(hash) < 40 {
		cl, _ := g.r.CommitObjects()
		_ = cl.ForEach(func(c *object.Commit) error {
			if c.Hash.String()[:len(hash)] == hash {
				commit = c
				cl.Close()
			}
			return nil
		})
	} else {
		commit, _ = g.r.CommitObject(plumbing.NewHash(hash))
	}

	if commit == nil {
		return nil, fmt.Errorf("commit not found")
	}

	parentCommit, errParent := commit.Parents().Next()
	if errParent != nil {
		if errParent == io.EOF {
			fl, _ := commit.Files()
			_ = fl.ForEach(func(f *object.File) error {
				files = append(files, f.Name)
				return nil
			})
		} else {
			return files, err
		}
	} else {
		patch, _ := parentCommit.Patch(commit)
		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()
			if from == nil {
				files = append(files, to.Path())
			} else if to == nil {
				files = append(files, from.Path())
			} else if from.Path() != to.Path() {
				files = append(files, to.Path())
			} else {
				files = append(files, from.Path())
			}
		}
	}
	return files, nil
}

var CreateGit = func(ctx *context.Context) (Manager, error) {
	r, err := git.PlainOpen(ctx.WorkingDir)
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()

	if err != nil {
		return nil, err
	}

	return Git{r: r, w: w}, nil
}

func GetStageFilesChanged(m Manager) []string {
	files := []string{}
	status, _ := m.Status()

	for path, fileStatus := range status {
		if fileStatus.Staging == git.Unmodified && fileStatus.Worktree == git.Unmodified {
			continue
		}

		if fileStatus.Staging == git.Renamed {
			path = fileStatus.Extra
		}
		files = append(files, path)
	}

	return files
}

func GetImagesNameAdded(ctx *context.Context, m Manager) ([]string, error) {
	names := []string{}

	status, _ := m.Status()
	afs := afero.Afero{Fs: ctx.FS}
	for path, fileStatus := range status {
		if !strings.Contains(path, loader.DataFilename) {
			continue
		}
		if fileStatus.Staging == git.Added || fileStatus.Worktree == git.Untracked {
			image := &types.Image{}

			content, err := afs.ReadFile(filepath.Join(ctx.WorkingDir, path))
			if err != nil {
				return names, fmt.Errorf("could not load file %s", path)
			}

			err = yaml.Unmarshal(content, image)
			if err != nil {
				return names, fmt.Errorf("could not parse %s with error : %s", path, err)
			}
			names = append(names, image.GetFullName())
		}
	}

	return names, nil
}

func GetImagesNameRemoved(m Manager) ([]string, error) {
	names := []string{}

	status, _ := m.Status()
	for path, fileStatus := range status {
		if fileStatus.Staging == git.Deleted && strings.Contains(path, loader.DataFilename) {
			image := &types.Image{}
			hash, errRevision := m.ResolveRevision(plumbing.Revision("HEAD^:" + path))
			if errRevision != nil {
				return names, errRevision
			}
			content, errContent := m.CommitFileContent(hash, path)
			if errContent != nil {
				return names, errContent
			}
			err := yaml.Unmarshal([]byte(content), image)
			if err != nil {
				return names, fmt.Errorf("could not parse removed %s with error : %s", path, err)
			}
			names = append(names, image.GetFullName())
		}
	}

	return names, nil
}

func CheckIfFileStaged(m Manager) bool {
	status, _ := m.Status()
	for _, fileStatus := range status {
		if slices.Contains([]git.StatusCode{git.Modified, git.Added, git.Renamed, git.Deleted}, fileStatus.Staging) {
			return true
		}
	}
	return false
}

func HasUntrackedFiles(m Manager) bool {
	status, _ := m.Status()

	for _, fileStatus := range status {
		if fileStatus.Worktree == git.Untracked {
			return true
		}
	}
	return false
}

func AddModifiedFilesToStage(m Manager) error {
	status, _ := m.Status()
	for filePath, fileStatus := range status {
		if fileStatus.Worktree != git.Modified && fileStatus.Worktree != git.Deleted {
			continue
		}
		err := AddFileToStage(m, filePath)
		if err != nil {
			return fmt.Errorf("add file %s fail with error: %v", filePath, err)
		}
	}
	return nil
}

func AddFileToStage(m Manager, path string) error {
	return m.AddWithOptions(&git.AddOptions{Path: path})
}
