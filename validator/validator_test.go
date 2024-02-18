package validator

import (
	"github.com/alexandreh2ag/mib/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidatePlatformParent(t *testing.T) {
	validate := New()
	tests := []struct {
		name       string
		image      types.Image
		wantErr    bool
		containErr string
	}{
		{
			name: "Success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(&tt.image)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.containErr)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidatePlatformParent_Success(t *testing.T) {
	validate := New()
	image := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Platforms: []string{"linux/amd64"}}
	images := types.Images{image}
	err := validate.Var(images, "dive")
	assert.NoError(t, err)
}

func TestValidatePlatformParent_SuccessWithParent(t *testing.T) {
	validate := New()
	image := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Platforms: []string{"linux/amd64"}, Parent: &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Platforms: []string{"linux/amd64", "linux/arm"}}}
	images := types.Images{image}
	err := validate.Var(images, "dive")
	assert.NoError(t, err)
}

func TestValidatePlatformParent_Fail(t *testing.T) {
	validate := New()

	image := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Platforms: []string{"linux/arm"}, Parent: &types.Image{Platforms: []string{"linux/adm64"}, ImageName: types.ImageName{Name: "foo", Tag: "0.1"}}}
	images := types.Images{image}
	err := validate.Var(images, "dive")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Key: '[0].Platforms' Error:Field validation for 'Platforms' failed on the 'platform-parent' tag")
}
