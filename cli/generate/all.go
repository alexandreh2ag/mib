package generate

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/loader"
	"github.com/alexandreh2ag/mib/template"
	"github.com/spf13/cobra"
)

func GetAllCmd(ctx *context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Generate all images readme",
		RunE:  GetAllRunFn(ctx),
	}
}

func GetAllRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		images, err := loader.LoadImages(ctx)
		if err != nil {
			return err
		}
		err = template.GenerateReadmeImages(ctx, images.GetAll())
		if err != nil {
			return err
		}

		return template.GenerateReadmeIndex(ctx, images, GetIndexReadmePath(ctx))
	}
}
