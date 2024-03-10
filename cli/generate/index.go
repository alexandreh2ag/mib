package generate

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/loader"
	"github.com/alexandreh2ag/mib/template"
	"github.com/spf13/cobra"
)

func GetIndexCmd(ctx *context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Generate index readme",
		RunE:  GetIndexRunFn(ctx),
	}
}

func GetIndexRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		images := loader.LoadImages(ctx)
		return template.GenerateReadmeIndex(ctx, images, GetIndexReadmePath(ctx))
	}
}
