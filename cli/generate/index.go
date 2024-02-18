package generate

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"
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
		panic(errors.New("implement me"))
		return nil
	}
}
