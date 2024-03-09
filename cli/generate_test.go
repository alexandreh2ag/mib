package cli

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGenerateCmd(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetGenerateCmd(ctx)

	assert.Equal(t, 3, len(cmd.Commands()))
}
