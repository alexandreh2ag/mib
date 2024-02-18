package main

import (
	"github.com/alexandreh2ag/mib/cli"
	"github.com/alexandreh2ag/mib/context"
)

func main() {
	ctx := context.DefaultContext()
	rootCmd := cli.GetRootCmd(ctx)
	rootCmd.AddCommand(
		cli.GetBuildCmd(ctx),
		cli.GetGenerateCmd(ctx),
		cli.GetListCmd(ctx),
		cli.GetCommitCmd(ctx),
		cli.GetVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
