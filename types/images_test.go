package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImages_MarkImagesToBuild(t *testing.T) {
	successOneImageToBuildWithChildFn := func() Images {
		child := &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, Path: "/test", HasLocalParent: true}
		parent := &Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/parent", Children: Images{child}}
		child.Parent = parent
		return Images{parent}
	}
	tests := []struct {
		name        string
		ims         Images
		pathToBuild []string
		fnCheck     func(t *testing.T, images Images)
	}{
		{
			name:        "SuccessEmptySlicesAndEmptyPaths",
			pathToBuild: []string{},
			fnCheck: func(t *testing.T, images Images) {
				assert.True(t, true)
			},
		},
		{
			name: "SuccessOneImageToBuild",
			ims: Images{
				&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/test"},
			},
			pathToBuild: []string{"/test"},
			fnCheck: func(t *testing.T, images Images) {
				assert.True(t, images[0].HasToBuild)
			},
		},
		{
			name: "SuccessOneImageNotToBuild",
			ims: Images{
				&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/test2"},
			},
			pathToBuild: []string{"/test"},
			fnCheck: func(t *testing.T, images Images) {
				assert.False(t, images[0].HasToBuild)
			},
		},
		{
			name: "SuccessOneImageChildToBuild",
			ims: Images{
				&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/parent", Children: Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, Path: "/test"}}},
			},
			pathToBuild: []string{"/test"},
			fnCheck: func(t *testing.T, images Images) {
				assert.False(t, images[0].HasToBuild)
				assert.True(t, images[0].Children[0].HasToBuild)
			},
		},
		{
			name:        "SuccessOneImageToBuildWithChild",
			ims:         successOneImageToBuildWithChildFn(),
			pathToBuild: []string{"/parent"},
			fnCheck: func(t *testing.T, images Images) {
				assert.True(t, images[0].HasToBuild, "Parent Image")
				assert.True(t, images[0].Children[0].HasToBuild, "Child Image")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ims.FlagImagesToBuild(tt.pathToBuild)
			if tt.fnCheck != nil {
				tt.fnCheck(t, tt.ims)
			}
		})
	}
}

func TestImages_GetImagesToBuild(t *testing.T) {
	tests := []struct {
		name string
		ims  Images
		want Images
	}{
		{
			name: "SuccessEmptySilce",
			ims:  Images{},
			want: Images{},
		},
		{
			name: "SuccessNoImageToBuild",
			ims:  Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}}},
			want: Images{},
		},
		{
			name: "SuccessImageToBuild",
			ims:  Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, HasToBuild: true}},
			want: Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, HasToBuild: true}},
		},
		{
			name: "SuccessImageToBuildWithChild",
			ims:  Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, HasToBuild: true, Children: Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, HasToBuild: true}}}},
			want: Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, HasToBuild: true, Children: Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, HasToBuild: true}}}, &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, HasToBuild: true}},
		},
		{
			name: "SuccessImageChildToBuild",
			ims:  Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Children: Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, HasToBuild: true}}}},
			want: Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, HasToBuild: true}},
		},
		{
			name: "SuccessOneImageToBuild",
			ims:  Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, HasToBuild: true}, &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}}},
			want: Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, HasToBuild: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.ims.GetImagesToBuild(), "GetImagesToBuild()")
		})
	}
}

func TestImages_FlagChanged(t *testing.T) {
	successOneImageToBuildWithChildFn := func() Images {
		child := &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, Path: "/test", HasLocalParent: true}
		parent := &Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/parent", Children: Images{child}}
		child.Parent = parent
		return Images{parent}
	}
	tests := []struct {
		name        string
		ims         Images
		pathToBuild []string
		fnCheck     func(t *testing.T, images Images)
	}{
		{
			name:        "SuccessEmptySlicesAndEmptyPaths",
			pathToBuild: []string{},
			fnCheck: func(t *testing.T, images Images) {
				assert.True(t, true)
			},
		},
		{
			name: "SuccessOneImageToBuild",
			ims: Images{
				&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/test"},
			},
			pathToBuild: []string{"/test/Dockerfile"},
			fnCheck: func(t *testing.T, images Images) {
				assert.True(t, images[0].HasToBuild)
			},
		},
		{
			name: "SuccessOneImageToBuildWithSubFileChanged",
			ims: Images{
				&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/test"},
			},
			pathToBuild: []string{"/test/rootfs/Dockerfile"},
			fnCheck: func(t *testing.T, images Images) {
				assert.True(t, images[0].HasToBuild)
			},
		},
		{
			name: "SuccessOneImageNotToBuild",
			ims: Images{
				&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/test2"},
			},
			pathToBuild: []string{"/test/Dockerfile"},
			fnCheck: func(t *testing.T, images Images) {
				assert.False(t, images[0].HasToBuild)
			},
		},
		{
			name: "SuccessOneImageChildToBuild",
			ims: Images{
				&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Path: "/parent", Children: Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, Path: "/test"}}},
			},
			pathToBuild: []string{"/test/Dockerfile"},
			fnCheck: func(t *testing.T, images Images) {
				assert.False(t, images[0].HasToBuild)
				assert.True(t, images[0].Children[0].HasToBuild)
			},
		},
		{
			name:        "SuccessOneImageToBuildWithChild",
			ims:         successOneImageToBuildWithChildFn(),
			pathToBuild: []string{"/parent/Dockerfile"},
			fnCheck: func(t *testing.T, images Images) {
				assert.True(t, images[0].HasToBuild, "Parent Image")
				assert.True(t, images[0].Children[0].HasToBuild, "Child Image")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ims.FlagChanged(tt.pathToBuild)
			if tt.fnCheck != nil {
				tt.fnCheck(t, tt.ims)
			}
		})
	}
}

func TestImages_GetAll_SuccessEmpty(t *testing.T) {
	images := Images{}
	want := Images{}
	assert.Equal(t, want, images.GetAll())
}

func TestImages_GetAll_SuccessWithoutChildren(t *testing.T) {
	images := Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}}, &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}}}
	want := Images{&Image{ImageName: ImageName{Name: "test", Tag: "0.1"}}, &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}}}
	assert.Equal(t, want, images.GetAll())
}

func TestImages_GetAll_SuccessWithout(t *testing.T) {
	imageChild2 := &Image{ImageName: ImageName{Name: "test", Tag: "0.1"}}
	imageChild1 := &Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Children: Images{imageChild2}}
	image1 := &Image{ImageName: ImageName{Name: "test", Tag: "0.1"}, Children: Images{imageChild1}}

	images := Images{image1, &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}}}
	want := Images{image1, imageChild1, imageChild2, &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}}}
	assert.Equal(t, want, images.GetAll())
}
