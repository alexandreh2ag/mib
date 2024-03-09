package cli

import (
	"errors"
	"github.com/alexandreh2ag/mib/context"
	"github.com/spf13/cobra"
)

const (
	gitStageAll   = "all"
	generateIndex = "index"
	generateImage = "image"
)

func GetCommitCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Commit all changes",
		RunE:  GetCommitRunFn(ctx),
	}

	cmd.Flags().BoolP(gitStageAll, "a", false, "Tell the command to automatically stage files that have been modified and deleted, but new files you have not told Git about are not affected.")
	cmd.Flags().Bool(generateIndex, true, "Generate index readme before add change")
	cmd.Flags().Bool(generateImage, true, "Generate images readme before add change")

	return cmd
}

func GetCommitRunFn(ctx *context.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return errors.New("implement me")
	}
}
