package docker

import (
	"errors"
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	mock_docker "github.com/alexandreh2ag/mib/mock/docker"
	"github.com/alexandreh2ag/mib/types"
	dockerApiTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCreateDockerBuilder_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	_ = os.Setenv("HOME", ctx.WorkingDir)
	_ = afero.WriteFile(ctx.FS, fmt.Sprintf("%s/.docker/config.json", ctx.WorkingDir), []byte("{\"auths\":{\"registry.example.com\":{\"auth\":\"dXNlcm5hbWU6cGFzc3dvcmQ=\"}}}"), 0644)
	got, err := CreateDockerBuilder(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateDockerBuilder_ErrorCreateClient(t *testing.T) {
	ctx := context.TestContext(nil)
	_ = os.Setenv(client.EnvOverrideHost, "error")
	got, err := CreateDockerBuilder(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to parse docker host `error`")
	assert.Nil(t, got)
	_ = os.Unsetenv(client.EnvOverrideHost)
}

func TestCreateDockerBuilder_ErrorGetAuth(t *testing.T) {
	ctx := context.TestContext(nil)
	_ = os.Setenv("HOME", ctx.WorkingDir)
	got, err := CreateDockerBuilder(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open /app/.docker/config.json: file does not exist")
	assert.Nil(t, got)
}

func TestBuilderDocker_Type(t *testing.T) {
	b := BuilderDocker{}
	assert.Equal(t, KeyBuilder, b.Type())
}

func TestBuilderDocker_Push_SuccessWithDockerHub(t *testing.T) {
	ctx := context.TestContext(nil)
	tag := "foo:0.1"
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"https://index.docker.io/v1/": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "https://index.docker.io/v1/", Username: "username", Password: "password"},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	clientDocker.EXPECT().ImagePush(gomock.Any(), tag, gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), nil)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.Push(tag)
	assert.NoError(t, err)
}

func TestBuilderDocker_Push_SuccessWithPrivateRegistry(t *testing.T) {
	ctx := context.TestContext(nil)
	tag := "registry.example.com/foo:0.1"
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry.example.com", Username: "username", Password: "password"},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	clientDocker.EXPECT().ImagePush(gomock.Any(), tag, gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), nil)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.Push(tag)
	assert.NoError(t, err)
}

func TestBuilderDocker_Push_ErrorParseTag(t *testing.T) {
	ctx := context.TestContext(nil)
	tag := "foo:0.1:wrong"
	auth := AuthConfig{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.Push(tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to format docker tag foo:0.1:wrong")
}

func TestBuilderDocker_Push_ErrorMissingAuth(t *testing.T) {
	ctx := context.TestContext(nil)
	tag := "registry.example.com/foo:0.1"
	auth := AuthConfig{AuthConfigs: map[string]registry.AuthConfig{}}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.Push(tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to find docker credential of registry.example.com.\n did you forget to docker login ?")
}

func TestBuilderDocker_PushImages_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com":  {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry.example.com", Username: "username", Password: "password"},
			"registry2.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry2.example.com", Username: "username", Password: "password"},
		},
	}
	image1Child := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo-bar", Tag: "0.1"}, HasToBuild: true}
	image1 := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo", Tag: "0.1"}, Alias: []types.ImageName{{Name: "registry2.example.com/foo", Tag: "0.1"}}, Children: types.Images{image1Child}, HasToBuild: true}
	images := types.Images{image1}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	gomock.InOrder(
		clientDocker.EXPECT().ImagePush(gomock.Any(), gomock.Eq(image1.GetFullName()), gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), nil),
		clientDocker.EXPECT().ImagePush(gomock.Any(), gomock.Eq(image1.Alias[0].GetFullName()), gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), nil),
		clientDocker.EXPECT().ImagePush(gomock.Any(), gomock.Eq(image1Child.GetFullName()), gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), nil),
	)

	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.PushImages(images)
	assert.NoError(t, err)
}

func TestBuilderDocker_PushImages_ErrorPush(t *testing.T) {
	ctx := context.TestContext(nil)
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry.example.com", Username: "username", Password: "password"},
		},
	}
	image1 := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo", Tag: "0.1"}, HasToBuild: true}
	images := types.Images{image1}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	gomock.InOrder(
		clientDocker.EXPECT().ImagePush(gomock.Any(), gomock.Eq(image1.GetFullName()), gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), errors.New("error")),
	)

	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.PushImages(images)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestBuilderDocker_PushImages_ErrorPushChild(t *testing.T) {
	ctx := context.TestContext(nil)
	auth := AuthConfig{
		AuthConfigs: map[string]registry.AuthConfig{
			"registry.example.com": {Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=", ServerAddress: "registry.example.com", Username: "username", Password: "password"},
		},
	}
	image1Child := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo-bar", Tag: "0.1"}, HasToBuild: true}
	image1 := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo", Tag: "0.1"}, Children: types.Images{image1Child}, HasToBuild: true}
	images := types.Images{image1}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	gomock.InOrder(
		clientDocker.EXPECT().ImagePush(gomock.Any(), gomock.Eq(image1.GetFullName()), gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), nil),
		clientDocker.EXPECT().ImagePush(gomock.Any(), gomock.Eq(image1Child.GetFullName()), gomock.Any()).Times(1).Return(io.NopCloser(strings.NewReader("")), errors.New("error")),
	)

	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.PushImages(images)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestBuilderDocker_Build_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	auth := AuthConfig{AuthConfigs: map[string]registry.AuthConfig{}}
	image := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo", Tag: "0.1"}, Alias: []types.ImageName{{Name: "registry2.example.com/foo", Tag: "0.1"}}}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	clientDocker.EXPECT().ImageBuild(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dockerApiTypes.ImageBuildResponse{Body: io.NopCloser(strings.NewReader(""))}, nil)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.Build(image)
	assert.NoError(t, err)
}

func TestBuilderDocker_BuildImages_Success(t *testing.T) {

	ctx := context.TestContext(nil)
	auth := AuthConfig{AuthConfigs: map[string]registry.AuthConfig{}}
	image1Child := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo-bar", Tag: "0.1"}, HasToBuild: true}
	image1 := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo", Tag: "0.1"}, Children: types.Images{image1Child}, HasToBuild: true}
	images := types.Images{image1}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	gomock.InOrder(
		clientDocker.EXPECT().ImageBuild(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dockerApiTypes.ImageBuildResponse{Body: io.NopCloser(strings.NewReader(""))}, nil),
		clientDocker.EXPECT().ImageBuild(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dockerApiTypes.ImageBuildResponse{Body: io.NopCloser(strings.NewReader(""))}, nil),
	)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.BuildImages(images)
	assert.NoError(t, err)
}

func TestBuilderDocker_BuildImages_ErrorBuild(t *testing.T) {
	ctx := context.TestContext(nil)
	auth := AuthConfig{AuthConfigs: map[string]registry.AuthConfig{}}
	image1Child := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo-bar", Tag: "0.1"}, HasToBuild: true}
	image1 := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo", Tag: "0.1"}, Children: types.Images{image1Child}, HasToBuild: true}
	images := types.Images{image1}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	gomock.InOrder(
		clientDocker.EXPECT().ImageBuild(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dockerApiTypes.ImageBuildResponse{Body: io.NopCloser(strings.NewReader(""))}, errors.New("error")),
	)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.BuildImages(images)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}

func TestBuilderDocker_BuildImages_ErrorBuildChild(t *testing.T) {
	ctx := context.TestContext(nil)
	auth := AuthConfig{AuthConfigs: map[string]registry.AuthConfig{}}
	image1Child := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo-bar", Tag: "0.1"}, HasToBuild: true}
	image1 := &types.Image{ImageName: types.ImageName{Name: "registry.example.com/foo", Tag: "0.1"}, Children: types.Images{image1Child}, HasToBuild: true}
	images := types.Images{image1}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientDocker := mock_docker.NewMockAPIClient(ctrl)
	gomock.InOrder(
		clientDocker.EXPECT().ImageBuild(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dockerApiTypes.ImageBuildResponse{Body: io.NopCloser(strings.NewReader(""))}, nil),
		clientDocker.EXPECT().ImageBuild(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dockerApiTypes.ImageBuildResponse{Body: io.NopCloser(strings.NewReader(""))}, errors.New("error")),
	)
	b := BuilderDocker{ctx: ctx, AuthConfig: &auth, client: clientDocker}
	err := b.BuildImages(images)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}
