package generate

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"

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
		panic(errors.New("implement me"))
		return nil
	}
}
