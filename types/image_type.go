package types

import (
	"fmt"
	"github.com/distribution/reference"
	"path"
	"strings"
)

const DockerHubUrl = "https://hub.docker.com/"

type ImageType int

const (
	UNDEFINED ImageType = iota
	OFFICIAL
	UNOFFICIAL
	REGISTRY
)

const DockerDomain = "docker.io"

func GetUrl(currentImage Image, destImage Image) (string, error) {
	switch GetImageType(destImage) {
	case OFFICIAL:
		return DockerHubUrl + "_/" + destImage.Name, nil
	case UNOFFICIAL:
		return DockerHubUrl + "r/" + destImage.Name, nil
	case REGISTRY:
		if destImage.RelativeDir == "" {
			return "", nil
		}
		depth := getDirDepth(currentImage.RelativeDir)
		return path.Join(strings.Repeat("../", depth), destImage.RelativeDir), nil
	default:
		return "", fmt.Errorf("unable to define ReadMeUrl (undefined Image type)")
	}

}

func getDirDepth(dir string) int {
	trimmed := strings.Trim(dir, "/")
	if trimmed == "" {
		return 0
	}
	joined := path.Join(trimmed)
	return len(strings.Split(joined, "/"))
}

func GetImageType(imageData Image) ImageType {
	ref, err := reference.ParseNormalizedNamed(imageData.Name)
	if err != nil {
		return UNDEFINED
	}
	if reference.Domain(ref) != DockerDomain {
		return REGISTRY
	} else if reference.Path(ref) == "library/"+imageData.Name {
		return OFFICIAL
	} else {
		return UNOFFICIAL
	}
}
