package cli

import (
	"github.com/alexandreh2ag/mib/context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBuildCmd(t *testing.T) {
	ctx := context.TestContext(nil)
	cmd := GetBuildCmd(ctx)

	assert.Equal(t, 2, len(cmd.Commands()))
}
