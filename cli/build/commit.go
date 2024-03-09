package build

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/cobra"
)

const Commit = "commit"

func GetCommitCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Build image for specific commit",
		RunE:  GetCommitRunFn(ctx),
	}

	cmd.Flags().String(Commit, "", "Commit sha")

	return cmd
}

func GetCommitRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return errors.New("implement me")
	}
}
