package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/alexandreh2ag/mib/container"
	mibContext "github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/exec"
	"github.com/alexandreh2ag/mib/types"
	typesContainers "github.com/alexandreh2ag/mib/types/container"
	"github.com/alexandreh2ag/mib/version"
	"github.com/distribution/reference"
	dockerApiTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
	"strings"
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

func (b BuilderDocker) BuildImages(images types.Images, pushImages bool) error {
	for _, image := range images {
		if image.HasToBuild {
			err := b.Build(image, pushImages)
			if err != nil {
				return fmt.Errorf("fail to build %s with error: %v", image.GetFullName(), err)
			}
		}
		if len(image.Children) > 0 {
			errChildren := b.BuildImages(image.Children, pushImages)
			if errChildren != nil {
				return errChildren
			}
		}
	}
	return nil
}

func (b BuilderDocker) Build(image *types.Image, pushImages bool) error {
	dockerCfg := b.ctx.Config.Build.Docker
	b.ctx.Logger.Info(fmt.Sprintf("Start building %s", image.GetFullName()))
	logger := b.ctx.Logger.With("image", image.Name)

	cmdArgs := []string{"build", "--progress", "plain"}

	if dockerCfg.CacheToEnable {
		cmdArgs = append(cmdArgs, "--cache-to", "type=inline,mode=max")
	}

	if dockerCfg.CacheFromEnable {
		cmdArgs = append(cmdArgs, "--cache-from", image.GetFullName())
	}

	if len(dockerCfg.BuildExtraOpts) > 0 {
		for optKey, optValue := range dockerCfg.BuildExtraOpts {
			cmdArgs = append(cmdArgs, fmt.Sprintf("--%s", optKey), optValue)
		}
	}

	labels := []string{
		fmt.Sprintf("%s=%s", "mib.version", version.GetFormattedVersion()),
	}
	argTags := sliceAddPrefixElement(image.GetNames(), "--tag")
	cmdArgs = append(cmdArgs, argTags...)
	argLabels := sliceAddPrefixElement(labels, "--label")
	cmdArgs = append(cmdArgs, argLabels...)

	if len(image.Platforms) > 0 {
		cmdArgs = append(cmdArgs, []string{"--platform", strings.Join(image.Platforms, ",")}...)
	}

	stdout := bytes.NewBuffer([]byte(""))
	stderr := bytes.NewBuffer([]byte(""))
	if pushImages {
		cmdArgs = append(cmdArgs, "--push")
	}
	cmdArgs = append(cmdArgs, ".")
	cmd := exec.NewCmd("docker", cmdArgs...)
	logger.Debug(fmt.Sprintf("command docker %s", cmdArgs))
	cmd.SetDir(image.Path)
	cmd.SetStdout(stdout)
	cmd.SetStderr(stderr)
	err := cmd.Run()
	logsLines := strings.Split(stderr.String(), "\n")
	for _, line := range logsLines {
		logger.Debug(line)
	}
	if err != nil {
		maxLine := 10
		offset := len(logsLines) - maxLine
		if len(logsLines) < maxLine {
			offset = 0
		}
		errorLines := logsLines[offset:]
		for _, line := range errorLines {
			logger.Error(line)
		}
		return err
	}

	b.ctx.Logger.Info(fmt.Sprintf("Finish building %s", image.GetFullName()))

	return nil
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
	b.ctx.Logger.Info(fmt.Sprintf("Start pushing %s", tag))
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
	pushResponse, errPush := b.client.ImagePush(context.Background(), tag, options)
	if errPush != nil {
		return errPush
	}
	defer func() {
		_ = pushResponse.Close()
	}()
	stringBuffer := bytes.NewBufferString("")
	termFd, isTerm := term.GetFdInfo(stringBuffer)
	_ = jsonmessage.DisplayJSONMessagesStream(pushResponse, stringBuffer, termFd, isTerm, nil)
	logger := b.ctx.Logger.With("image", tag)
	for _, line := range strings.Split(stringBuffer.String(), "\n") {
		logger.Debug(line)
	}
	b.ctx.Logger.Info(fmt.Sprintf("Finish pushing %s", tag))
	return err
}

func sliceAddPrefixElement(list []string, prefix string) []string {
	result := []string{}
	for _, s := range list {
		result = append(result, []string{prefix, s}...)
	}
	return result
}
