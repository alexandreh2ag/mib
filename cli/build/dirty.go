package build

import (
	"github.com/alexandreh2ag/mib/container/docker"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/git"
	"github.com/alexandreh2ag/mib/loader"

	"github.com/spf13/cobra"
)

func GetDirtyCmd(ctx *context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "dirty",
		Short: "Build image with change not committed",
		RunE:  GetDirtyRunFn(ctx),
	}
}

func GetDirtyRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		_, _ = cmd.Flags().GetBool(DryRun)
		pushImages, _ := cmd.Flags().GetBool(PushImages)
		builder := ctx.Builders.GetInstance(docker.KeyBuilder)
		gitManager, errGit := git.CreateGit(ctx)
		if errGit != nil {
			return errGit
		}
		images := loader.LoadImages(ctx)
		filesChanged := git.GetStageFilesChanged(gitManager)
		images.FlagChanged(loader.RemoveExtExcludePath(ctx.WorkingDir, ctx.Config.Build.ExtensionExclude, filesChanged))
		errBuild := builder.BuildImages(images)
		if errBuild != nil {
			return errBuild
		}

		if pushImages {
			errPush := builder.PushImages(images)
			if errPush != nil {
				return errPush
			}
		}

		return nil
	}
}
