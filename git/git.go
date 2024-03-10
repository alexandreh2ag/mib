package git

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/go-git/go-git/v5"
)

const (
	Added    = "+"
	Updated  = "*"
	Removed  = "-"
	NoUpdate = ""
)

type Worktree interface {
	Status() (git.Status, error)
}

var GetWorktree = func(ctx *context.Context) (Worktree, error) {
	r, err := git.PlainOpen(ctx.WorkingDir)
	if err != nil {
		return nil, err
	}
	workTree, err := r.Worktree()

	if err != nil {
		return workTree, err
	}

	return workTree, nil
}

func GetStageFilesChanged(ctx *context.Context) ([]string, error) {
	files := []string{}
	workTree, err := GetWorktree(ctx)
	if err != nil {
		return files, err
	}
	status, _ := workTree.Status()

	for path, fileStatus := range status {
		if fileStatus.Staging == git.Unmodified && fileStatus.Worktree == git.Unmodified {
			continue
		}

		if fileStatus.Staging == git.Renamed {
			path = fileStatus.Extra
		}
		files = append(files, path)
	}

	return files, nil
}
