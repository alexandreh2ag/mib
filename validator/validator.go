package validator

import (
	"github.com/alexandreh2ag/mib/types"
	"github.com/go-playground/validator/v10"
	"slices"
)

const (
	PlatformParent = "platform-parent"
)

func New(options ...validator.Option) *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	_ = validate.RegisterValidation(PlatformParent, ValidateImagePlatformParent())
	return validate
}

func ValidateImagePlatformParent() func(level validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		image := fl.Parent().Interface().(types.Image)

		if image.Parent != nil && len(image.Parent.Platforms) > 0 {
			for _, platform := range image.Platforms {
				if !slices.Contains(image.Parent.Platforms, platform) {
					return false
				}
			}
		}

		return true
	}
}
