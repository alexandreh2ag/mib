package container

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/types/container"
)

var BuilderFnFactory = map[string]CreateBuilderFn{}

type CreateBuilderFn = func(ctx *context.Context) (container.BuilderImage, error)
