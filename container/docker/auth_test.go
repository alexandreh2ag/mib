package docker

import (
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/docker/docker/api/types/registry"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetAuthConfig_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	_ = os.Setenv("HOME", ctx.WorkingDir)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/.docker/config.json", ctx.WorkingDir), []byte("{\"auths\":{\"registry.example.com\":{\"auth\":\"dXNlcm5hbWU6cGFzc3dvcmQ=\"}}}"), 0644)
	want := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry.example.com", Username: "username", Password: "password"},
		},
	}
	got, err := GetAuthConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetAuthConfig_ErrorReadFile(t *testing.T) {
	ctx := context.TestContext(nil)
	_ = os.Setenv("HOME", ctx.WorkingDir)
	want := AuthConfig{}
	got, err := GetAuthConfig(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open /app/.docker/config.json: file does not exist")
	assert.Equal(t, want, got)
}

func TestGetAuthConfig_ErrorUnmarshal(t *testing.T) {
	ctx := context.TestContext(nil)
	_ = os.Setenv("HOME", ctx.WorkingDir)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/.docker/config.json", ctx.WorkingDir), []byte("{]"), 0644)
	want := AuthConfig{}
	got, err := GetAuthConfig(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character ']' looking for beginning of object key string")
	assert.Equal(t, want, got)
}

func TestAuthConfig_GetAuthConfigs_Success(t *testing.T) {
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com":  {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ="},
			"registry2.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ="},
		},
	}
	want := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com":  {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry.example.com", Username: "username", Password: "password"},
			"registry2.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry2.example.com", Username: "username", Password: "password"},
		},
	}
	err := auth.GetAuthConfigs()
	assert.NoError(t, err)
	assert.Equal(t, want, auth)
}

func TestAuthConfig_GetAuthConfigs_ErrorDecode(t *testing.T) {
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ"},
		},
	}
	want := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ"},
		},
	}
	err := auth.GetAuthConfigs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot decode base64 string from .docker/config.json")
	assert.Equal(t, want, auth)
}

func TestAuthConfig_GetAuthConfigs_ErrorSplitMoreElement(t *testing.T) {
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ6d3Jvbmc="},
		},
	}
	want := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ6d3Jvbmc="},
		},
	}
	err := auth.GetAuthConfigs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base64 auth filed is malformed in .docker/config.json for registry.example.com registry")
	assert.Equal(t, want, auth)
}
