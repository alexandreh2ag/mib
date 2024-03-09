package loader

import (
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/types"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

var DataFilename = "dbp.yml"

func LoadImages(ctx *context.Context) types.Images {
	images := types.Images{}
	afs := &afero.Afero{Fs: ctx.FS}
	_ = afs.Walk(ctx.WorkingDir, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			ctx.Logger.Error(fmt.Sprintf("Cannot walk there : %s.", err))
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		if DataFilename == fi.Name() {
			ctx.Logger.Debug(fmt.Sprintf("Try to load dbp file : %s", path.Dir(fp)))
			image, loadErr := LoadImage(ctx, fp)
			if loadErr != nil {
				ctx.Logger.Error(fmt.Sprintf("Error occured : %s.", loadErr))
			} else {
				images = append(images, &image)
			}
		}
		return nil
	})
	return orderDependencyImages(images)
}

func orderDependencyImages(imagesToSort types.Images) types.Images {
	imagesSorted := types.Images{}
	for _, mainImageData := range imagesToSort {
		for _, imageData := range imagesToSort {
			if imageData.Parent != nil {
				isParent := mainImageData.Name == imageData.Parent.Name && mainImageData.Tag == imageData.Parent.Tag
				if isParent {
					imageData.HasParentToBuild = true
					imageData.HasLocalParent = true
					imageData.Parent = mainImageData
					mainImageData.Children = append(mainImageData.Children, imageData)
				}
			}
		}
	}

	for _, imageData := range imagesToSort {
		if !imageData.HasParentToBuild {
			imagesSorted = append(imagesSorted, imageData)
		}
	}

	return imagesSorted
}

func LoadImage(ctx *context.Context, path string) (types.Image, error) {
	afs := &afero.Afero{Fs: ctx.FS}
	image := types.Image{}
	image.Path = filepath.Dir(path)
	relativePath, _ := filepath.Rel(ctx.WorkingDir, image.Path)
	image.RelativeDir = relativePath
	content, err := afs.ReadFile(path)
	if err != nil {
		return image, fmt.Errorf("could not load file %s", path)
	}

	err = yaml.Unmarshal(content, &image)
	if err != nil {
		return image, fmt.Errorf("could not parse %s with error : %s", path, err)
	}

	err = findParentImage(ctx, &image)
	if err != nil {
		return image, err
	}

	ctx.Logger.Debug(fmt.Sprintf("Image %s loaded.", image.GetFullName()))

	return image, nil
}

func findParentImage(ctx *context.Context, image *types.Image) error {
	afs := &afero.Afero{Fs: ctx.FS}
	parentImage := types.Image{}
	dockerFileContent, err := afs.ReadFile(filepath.Join(image.Path, "/Dockerfile"))

	if err != nil {
		return fmt.Errorf("could not read dockerFile of image %s", image.GetFullName())
	}

	rgx := regexp.MustCompile(`FROM\s*(.*):(.*)`)
	regexResult := rgx.FindStringSubmatch(string(dockerFileContent))

	if len(regexResult) > 3 || len(regexResult) == 0 {
		return fmt.Errorf("dockerFile of %s contains zero or multiple FROM", image.GetFullName())
	}

	parentImage.Name = regexResult[1]
	parentImage.Tag = regexResult[2]
	image.Parent = &parentImage

	return nil
}
