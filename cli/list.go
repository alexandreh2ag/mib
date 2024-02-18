package cli

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/loader"
	"github.com/alexandreh2ag/mib/printer"
	"github.com/spf13/cobra"
)

func GetListCmd(ctx *context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all images of directory",
		RunE:  GetListRunFn(ctx),
	}
}

func GetListRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		images, err := loader.LoadImages(ctx)
		if err != nil {
			return err
		}
		if len(images) > 0 {
			cmd.Println(printer.DisplayImagesTree(images))
		} else {
			cmd.Println("No images loaded")
		}

		return nil
	}
}
