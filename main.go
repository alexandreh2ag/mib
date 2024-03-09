package main

import (
	"github.com/alexandreh2ag/mib/cli"
	"github.com/alexandreh2ag/mib/context"
)

func main() {
	ctx := context.DefaultContext()
	rootCmd := cli.GetRootCmd(ctx)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
