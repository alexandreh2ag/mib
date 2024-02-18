package cli

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"
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
		panic(errors.New("implement me"))

		return nil
	}
}
