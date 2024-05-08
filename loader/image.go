package loader

import (
	"errors"
	"fmt"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/types"
	validatorMIB "github.com/alexandreh2ag/mib/validator"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var DataFilename = "mib.yml"

func LoadImages(ctx *context.Context) (types.Images, error) {
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
			ctx.Logger.Debug(fmt.Sprintf("Try to load mib file : %s", path.Dir(fp)))
			image, loadErr := LoadImage(ctx, fp)
			if loadErr != nil {
				ctx.Logger.Error(fmt.Sprintf("Error occured : %s.", loadErr))
			} else {
				images = append(images, &image)
			}
		}
		return nil
	})
	imagesOrdered := orderDependencyImages(images)
	validate := validatorMIB.New()
	err := validate.Var(imagesOrdered, "dive")
	if err != nil {
		var validationErrors validator.ValidationErrors
		switch {
		case errors.As(err, &validationErrors):
			for _, validationError := range validationErrors {
				ctx.Logger.Error(fmt.Sprintf("%v", reformatValidatorError(imagesOrdered, validationError)))
			}
			return imagesOrdered, errors.New("images configuration file is not valid")
		default:
			return imagesOrdered, err
		}
	}
	return imagesOrdered, nil
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
					if len(imageData.Platforms) == 0 && len(mainImageData.Platforms) > 0 {
						imageData.Platforms = mainImageData.Platforms
					}
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

func RemoveExtExcludePath(workingDir string, extensionExclude string, filesUpdated []string) []string {
	pathsChanged := []string{}
	for _, file := range filesUpdated {
		if extensionExclude == "" || !slices.Contains(strings.Split(extensionExclude, ","), filepath.Ext(file)) {
			fileChangedAbsolute, _ := filepath.Abs(filepath.Join(workingDir, file))
			pathsChanged = append(pathsChanged, fileChangedAbsolute)
		}
	}
	return pathsChanged
}

func reformatValidatorError(images types.Images, err validator.FieldError) string {
	errNamespace := err.Namespace()
	regex := regexp.MustCompile(`^\[(?P<index>\d)]\.`)
	regexChildren := regexp.MustCompile(`(Children\[(\d)*])\.`)
	match := regex.FindStringSubmatch(errNamespace)
	if len(match) >= 2 {
		replaceStr := match[0]
		indexImage, _ := strconv.Atoi(match[1])
		currentImage := images[indexImage]
		imagePath := currentImage.Path
		matchChildren := regexChildren.FindAllStringSubmatch(errNamespace, -1)
		for _, matchChild := range matchChildren {
			replaceStr += matchChild[0]
			indexImage, _ = strconv.Atoi(match[1])
			currentImage = currentImage.Children[indexImage]
			imagePath = currentImage.Path
		}

		return strings.Replace(err.Error(), replaceStr, imagePath+" ", 1)
	}
	return err.Error()
}
