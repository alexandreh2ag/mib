package types

import "strings"

type Images []*Image

func (ims Images) FlagImagesToBuild(pathToBuild []string) {
	for _, path := range pathToBuild {
		for _, image := range ims {
			if path == image.Path {
				image.HasToBuild = true
			} else if image.HasLocalParent && image.Parent.HasToBuild {
				image.HasToBuild = true
			}
			image.Children.FlagImagesToBuild(pathToBuild)
		}
	}
}
func (ims Images) GetAll() Images {
	images := Images{}
	for _, image := range ims {
		images = append(images, image)
		images = append(images, image.Children.GetAll()...)

	}
	return images
}

func (ims Images) GetImagesToBuild() Images {
	images := Images{}

	for _, image := range ims {
		if image.HasToBuild {
			images = append(images, image)
		}
		images = append(images, image.Children.GetImagesToBuild()...)
	}
	return images
}

func (ims Images) FlagChanged(pathToBuild []string) {
	for _, path := range pathToBuild {
		for _, image := range ims {
			if strings.Contains(path, image.Path) {
				image.HasToBuild = true
			} else if image.HasLocalParent && image.Parent.HasToBuild {
				image.HasToBuild = true
			}
			image.Children.FlagChanged(pathToBuild)
		}
	}
}
