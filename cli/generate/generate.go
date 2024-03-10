package generate

import (
	"github.com/alexandreh2ag/mib/context"
	"path/filepath"
)

func GetIndexReadmePath(ctx *context.Context) string {
	return filepath.Join(ctx.WorkingDir, "README.md")
}
