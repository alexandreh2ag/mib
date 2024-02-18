package generate

import (
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetIndexReadmePath(t *testing.T) {
	ctx := context.DefaultContext()

	assert.Equalf(t, fmt.Sprintf("%s/README.md", ctx.WorkingDir), GetIndexReadmePath(ctx), "GetIndexReadmePath(%v)", ctx)
}
