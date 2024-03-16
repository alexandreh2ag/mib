package cli

import (
	"errors"
	"fmt"
	"github.com/alexandreh2ag/mib/cli/generate"
	"github.com/alexandreh2ag/mib/context"
	mibGit "github.com/alexandreh2ag/mib/git"
	"github.com/alexandreh2ag/mib/loader"
	"github.com/alexandreh2ag/mib/template"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

const (
	gitStageAll   = "all"
	generateIndex = "index"
	generateImage = "image"
)

func GetCommitCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Commit all changes",
		RunE:  GetCommitRunFn(ctx),
	}

	cmd.Flags().BoolP(gitStageAll, "a", false, "Tell the command to automatically stage files that have been modified and deleted, but new files you have not told Git about are not affected.")
	cmd.Flags().Bool(generateIndex, false, "Generate index readme before add change")
	cmd.Flags().Bool(generateImage, false, "Generate images readme before add change")

	return cmd
}

func GetCommitRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		stageChange, _ := cmd.Flags().GetBool(gitStageAll)
		genIndex, _ := cmd.Flags().GetBool(generateIndex)
		genImage, _ := cmd.Flags().GetBool(generateImage)
		ctx.Logger.Error(fmt.Sprintf("Stage Changed: %v; Generate index readme: %v; Generate images readme: %v", stageChange, genIndex, genImage))

		gitManager, errGit := mibGit.CreateGit(ctx)
		if errGit != nil {
			return errGit
		}

		if mibGit.HasUntrackedFiles(gitManager) {
			ctx.Logger.Warn("Some files are untracked. this may provoc inconsistent in build.")
		}

		images := loader.LoadImages(ctx)
		filesChanged := mibGit.GetStageFilesChanged(gitManager)

		images.FlagChanged(loader.RemoveExtExcludePath(ctx.WorkingDir, ctx.Config.Build.ExtensionExclude, filesChanged))

		nameImagesAdded, errAdded := mibGit.GetImagesNameAdded(ctx, gitManager)
		if errAdded != nil {
			return errAdded
		}

		nameImagesRemoved, errRemoved := mibGit.GetImagesNameRemoved(gitManager)
		if errRemoved != nil {
			return errRemoved
		}

		nameImagesUpdated := sliceDifference(images.GetImagesToBuild().GetAllNames(false), nameImagesAdded)

		if stageChange {
			err := mibGit.AddModifiedFilesToStage(gitManager)
			if err != nil {
				return err
			}
		}

		if !mibGit.CheckIfFileStaged(gitManager) {
			return errors.New("no change detected in stage")
		}

		ctx.Logger.Info(fmt.Sprintf("Images added %d, modified %d, removed %d", len(nameImagesAdded), len(nameImagesUpdated), len(nameImagesRemoved)))
		if genIndex && len(filesChanged) > 0 {
			errGenIndex := template.GenerateReadmeIndex(ctx, images, generate.GetIndexReadmePath(ctx))
			if errGenIndex != nil {
				return errGenIndex
			}
			_ = mibGit.AddFileToStage(gitManager, "README.md")
		}

		if genImage {
			errGenImage := template.GenerateReadmeImages(ctx, images.GetImagesToBuild())
			if errGenImage != nil {
				return errGenImage
			}
			for _, image := range images.GetImagesToBuild() {
				path, _ := filepath.Rel(ctx.WorkingDir, image.Path)
				_ = mibGit.AddFileToStage(gitManager, filepath.Join(path, "README.md"))
			}
		}
		message, errMessage := getCommitMessage(nameImagesAdded, nameImagesUpdated, nameImagesRemoved)
		if errMessage != nil {
			return errMessage
		}

		hash, errCommit := gitManager.CreateCommit(message, &git.CommitOptions{All: false, AllowEmptyCommits: false})
		if errCommit != nil {
			return errCommit
		}
		ctx.Logger.Info(fmt.Sprintf("commit created with hash: %s", hash.String()))
		return nil
	}
}

func getCommitMessage(nameImagesAdded, nameImagesUpdated, nameImagesRemoved []string) (string, error) {
	if len(nameImagesAdded) == 0 && len(nameImagesUpdated) == 0 && len(nameImagesRemoved) == 0 {
		return "", fmt.Errorf("no change detected to generate commit message")
	}
	message := fmt.Sprintf(
		"This commit add %d, update %d and %d removed images\n\n",
		len(nameImagesAdded),
		len(nameImagesUpdated),
		len(nameImagesRemoved),
	)
	nameImagesAdded = sliceAddPrefix(nameImagesAdded, mibGit.Added+" ")
	message += strings.Join(nameImagesAdded, "\n") + "\n"

	nameImagesUpdated = sliceAddPrefix(nameImagesUpdated, mibGit.Updated+" ")
	message += strings.Join(nameImagesUpdated, "\n") + "\n"

	nameImagesRemoved = sliceAddPrefix(nameImagesRemoved, mibGit.Removed+" ")
	message += strings.Join(nameImagesRemoved, "\n") + "\n"

	return message, nil
}

func sliceAddPrefix(s []string, prefix string) []string {
	for i, val := range s {
		s[i] = fmt.Sprintf("%s%s", prefix, val)
	}
	return s
}

func sliceDifference(s1 []string, s2 []string) []string {
	difference := []string{}
	found := false
	for _, val1 := range s1 {
		for _, val2 := range s2 {
			if val1 == val2 {
				found = true
				break
			}
		}
		if !found {
			difference = append(difference, val1)
		}
	}
	return difference
}
