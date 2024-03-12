package cli

import (
	"github.com/alexandreh2ag/mib/cli/build"
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/cobra"
)

func GetBuildCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "build sub commands",
	}
	cmd.PersistentFlags().Bool(build.PushImages, false, "Push image to registry")
	cmd.PersistentFlags().BoolP(build.DryRun, "d", false, "Dry run")

	cmd.AddCommand(build.GetDirtyCmd(ctx))
	cmd.AddCommand(build.GetCommitCmd(ctx))

	return cmd
}
