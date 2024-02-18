package generate

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/git"
	"github.com/alexandreh2ag/mib/loader"
	"github.com/alexandreh2ag/mib/template"

	"github.com/spf13/cobra"
)

func GetDirtyCmd(ctx *context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "dirty",
		Short: "Generate image readme with change not committed",
		RunE:  GetDirtyRunFn(ctx),
	}
}

func GetDirtyRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		gitManager, errGit := git.CreateGit(ctx)
		if errGit != nil {
			return errGit
		}
		images, err := loader.LoadImages(ctx)
		if err != nil {
			return err
		}

		filesChanged := git.GetStageFilesChanged(gitManager)
		images.FlagChanged(loader.RemoveExtExcludePath(ctx.WorkingDir, ctx.Config.Build.ExtensionExclude, filesChanged))
		return template.GenerateReadmeImages(ctx, images.GetImagesToBuild())
	}
}
