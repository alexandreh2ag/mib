package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/docker/docker/api/types/registry"
	"github.com/spf13/afero"
	"os"
	"path"
	"strings"
)

type AuthConfig struct {
	AuthConfigs map[string]registry.AuthConfig `json:"auths,omitempty"`
	HttpHeaders struct {
		UserAgent string `json:"User-Agent,omitempty"`
	}
}

func (ac *AuthConfig) GetAuthConfigs() error {
	for hostname, config := range ac.AuthConfigs {
		data, err := base64.StdEncoding.DecodeString(config.Auth)
		if err != nil {
			return fmt.Errorf("cannot decode base64 string from .docker/config.json")
		}

		usernamePassword := strings.Split(string(data), ":")
		if len(usernamePassword) != 2 {
			return fmt.Errorf("base64 auth filed is malformed in .docker/config.json for %s registry", hostname)
		}

		ac.AuthConfigs[hostname] = registry.AuthConfig{
			Username:      usernamePassword[0],
			Password:      usernamePassword[1],
			Auth:          config.Auth,
			ServerAddress: hostname,
		}
	}

	return nil
}

func GetAuthConfig(ctx *context.Context) (AuthConfig, error) {
	afs := &afero.Afero{Fs: ctx.FS}
	authConfig := AuthConfig{}
	configFile, err := afs.ReadFile(path.Join(os.Getenv("HOME"), ".docker", "config.json"))

	if err != nil {
		return authConfig, err
	}
	err = json.Unmarshal(configFile, &authConfig)

	if err != nil {
		return authConfig, err
	}

	err = authConfig.GetAuthConfigs()

	return authConfig, err
}
