package printer

import (
	"github.com/alexandreh2ag/mib/types"
	"github.com/stretchr/testify/assert"
	"github.com/xlab/treeprint"
	"testing"
)

func Test_displayImages(t *testing.T) {
	tests := []struct {
		name   string
		images types.Images
		want   string
	}{
		{
			name:   "SuccessEmpty",
			images: types.Images{},
			want:   ".\n",
		},
		{
			name:   "SuccessOneImage",
			images: types.Images{&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}}},
			want:   ".\n└── foo:0.1\n",
		},
		{
			name:   "SuccessOneImageHasToBuild",
			images: types.Images{&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}, HasToBuild: true}},
			want:   ".\n└── foo:0.1\n",
		},
		{
			name:   "SuccessTwoImagesOneLevel",
			images: types.Images{&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}}, &types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.2"}}},
			want:   ".\n├── foo:0.1\n└── foo:0.2\n",
		},
		{
			name:   "SuccessTwoImagesTwoLevel",
			images: types.Images{&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}, Children: types.Images{&types.Image{ImageName: types.ImageName{Name: "foo-bar", Tag: "0.1"}}}}, &types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.2"}}},
			want:   ".\n├── foo:0.1\n│   └── foo-bar:0.1\n└── foo:0.2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := treeprint.New()
			displayImages(tt.images, tree)
			assert.Equal(t, tt.want, tree.String())
		})
	}
}

func TestDisplayImagesTree(t *testing.T) {
	tests := []struct {
		name   string
		images types.Images
		want   string
	}{
		{
			name:   "SuccessEmpty",
			images: types.Images{},
			want:   ".\n",
		},
		{
			name:   "SuccessOneImage",
			images: types.Images{&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}}},
			want:   ".\n└── foo:0.1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DisplayImagesTree(tt.images), "DisplayImagesTree(%v)", tt.images)
		})
	}
}
