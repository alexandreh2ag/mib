package build

import (
	"github.com/alexandreh2ag/mib/container/docker"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/git"
	"github.com/alexandreh2ag/mib/loader"
	"github.com/spf13/cobra"
)

const Commit = "commit"

func GetCommitCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Build image for specific commit",
		RunE:  GetCommitRunFn(ctx),
	}

	cmd.Flags().String(Commit, "", "Commit sha")

	return cmd
}

func GetCommitRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		pushImages, _ := cmd.Flags().GetBool(PushImages)
		commitHash, _ := cmd.Flags().GetString(Commit)
		builder := ctx.Builders.GetInstance(docker.KeyBuilder)
		gitManager, errCreateGit := git.CreateGit(ctx)
		if errCreateGit != nil {
			return errCreateGit
		}
		images, err := loader.LoadImages(ctx)
		if err != nil {
			return err
		}
		filesChanged, errGetChanged := gitManager.GetCommitFilesChanged(commitHash)
		if errGetChanged != nil {
			return errGetChanged
		}

		images.FlagChanged(loader.RemoveExtExcludePath(ctx.WorkingDir, ctx.Config.Build.ExtensionExclude, filesChanged))
		errBuild := builder.BuildImages(images, pushImages)
		if errBuild != nil {
			return errBuild
		}

		return nil
	}
}
