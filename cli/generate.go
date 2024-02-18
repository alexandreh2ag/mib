package cli

import (
	"github.com/alexandreh2ag/mib/cli/generate"
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/cobra"
)

func GetGenerateCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "generate sub commands",
	}
	cmd.AddCommand(generate.GetIndexCmd(ctx))
	cmd.AddCommand(generate.GetAllCmd(ctx))
	cmd.AddCommand(generate.GetDirtyCmd(ctx))

	return cmd
}
