package generate

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"
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
		return errors.New("implement me")
	}
}
