package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/alexandreh2ag/mib/container"
	mibContext "github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/types"
	typesContainers "github.com/alexandreh2ag/mib/types/container"
	"github.com/alexandreh2ag/mib/version"
	"github.com/distribution/reference"
	dockerApiTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

const (
	KeyBuilder = "docker"
	AuthUrl    = "https://index.docker.io/v1/"
	Domain     = "docker.io"
)

func init() {
	container.BuilderFnFactory[KeyBuilder] = CreateDockerBuilder
}

func CreateDockerBuilder(ctx *mibContext.Context) (typesContainers.BuilderImage, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	authConfigData, errAuth := GetAuthConfig(ctx)
	if errAuth != nil {
		return nil, errAuth
	}

	return &BuilderDocker{ctx: ctx, client: cli, AuthConfig: &authConfigData}, nil
}

var _ typesContainers.BuilderImage = &BuilderDocker{}

type BuilderDocker struct {
	ctx        *mibContext.Context
	client     client.APIClient
	AuthConfig *AuthConfig
}

func (b BuilderDocker) Type() string {
	return KeyBuilder
}

func (b BuilderDocker) BuildImages(images types.Images) error {
	for _, image := range images {
		if image.HasToBuild {
			err := b.Build(image)
			if err != nil {
				return fmt.Errorf("fail to build %s with error: %v", image.GetFullName(), err)
			}
		}
		if len(image.Children) > 0 {
			errChildren := b.BuildImages(image.Children)
			if errChildren != nil {
				return errChildren
			}
		}
	}
	return nil
}

func (b BuilderDocker) Build(image *types.Image) error {
	b.ctx.Logger.Debug(fmt.Sprintf("Start building %s", image.GetFullName()))
	dockerBuildContext, err := archive.TarWithOptions(image.Path, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("error when build tar option for %s with error: %v", image.GetFullName(), err)
	}
	defer func() {
		_ = dockerBuildContext.Close()
	}()

	args := map[string]*string{}
	options := dockerApiTypes.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     false,
		Tags:           image.GetNames(),
		BuildArgs:      args,
		AuthConfigs:    b.AuthConfig.AuthConfigs,
		Labels: map[string]string{
			"mib.version": version.GetFormattedVersion(),
		},
	}
	_, err = b.client.ImageBuild(context.Background(), dockerBuildContext, options)
	b.ctx.Logger.Debug(fmt.Sprintf("Finish building %s", image.GetFullName()))
	return err
}

func (b BuilderDocker) PushImages(images types.Images) error {
	for _, image := range images {
		if image.HasToBuild {
			for _, tag := range image.GetNames() {
				err := b.Push(tag)
				if err != nil {
					return err
				}
			}
		}
		if len(image.Children) > 0 {
			errChildren := b.PushImages(image.Children)
			if errChildren != nil {
				return errChildren
			}
		}
	}
	return nil
}

func (b BuilderDocker) Push(tag string) error {
	b.ctx.Logger.Debug(fmt.Sprintf("Start pushing %s", tag))
	ref, err := reference.ParseNormalizedNamed(tag)
	if err != nil {
		return fmt.Errorf("unable to format docker tag %s", tag)
	}
	authKey := reference.Domain(ref)

	if reference.Domain(ref) == Domain {
		authKey = AuthUrl
	}

	if _, ok := b.AuthConfig.AuthConfigs[authKey]; !ok {
		return fmt.Errorf("unable to find docker credential of %s.\n did you forget to docker login ?", reference.Domain(ref))
	}
	authString, _ := json.Marshal(b.AuthConfig.AuthConfigs[authKey])

	options := dockerApiTypes.ImagePushOptions{
		RegistryAuth: base64.URLEncoding.EncodeToString(authString),
		All:          false,
	}
	_, err = b.client.ImagePush(context.Background(), tag, options)
	b.ctx.Logger.Debug(fmt.Sprintf("Finish pushing %s", tag))
	return err
}
