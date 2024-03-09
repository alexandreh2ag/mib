package types

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
