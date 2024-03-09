package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImageName_GetFullName(t *testing.T) {
	type fields struct {
		Name string
		Tag  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "SuccessEmpty",
			want: ":",
		},
		{
			name: "SuccessValidName",
			fields: fields{
				Name: "test",
				Tag:  "0.1",
			},
			want: "test:0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := Image{
				ImageName: ImageName{
					Name: tt.fields.Name,
					Tag:  tt.fields.Tag,
				},
			}
			assert.Equalf(t, tt.want, im.GetFullName(), "GetFullName()")
		})
	}
}

func TestImage_GetParents(t *testing.T) {

	tests := []struct {
		name   string
		Parent *Image
		want   Images
	}{
		{
			name: "SuccessNoParent",
		},
		{
			name:   "SuccessGotParentOneLevel",
			Parent: &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}},
			want:   Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}}},
		},
		{
			name:   "SuccessGotParentTwoLevel",
			Parent: &Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, Parent: &Image{ImageName: ImageName{Name: "test", Tag: "0.2"}}},
			want:   Images{&Image{ImageName: ImageName{Name: "foo", Tag: "0.1"}, Parent: &Image{ImageName: ImageName{Name: "test", Tag: "0.2"}}}, &Image{ImageName: ImageName{Name: "test", Tag: "0.2"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := Image{
				Parent: tt.Parent,
			}
			assert.Equalf(t, tt.want, im.GetParents(), "GetParents()")
		})
	}
}

func TestImage_GetAllEnvVar(t *testing.T) {
	type fields struct {
		Parent       *Image
		EnvVariables map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "SuccessNoEnv",
			want: map[string]string{},
		},
		{
			name:   "SuccessEnvWithNoParent",
			fields: fields{EnvVariables: map[string]string{"test": "test"}},
			want:   map[string]string{"test": "test"},
		},
		{
			name:   "SuccessEnvWithParent",
			fields: fields{EnvVariables: map[string]string{"test": "test"}, Parent: &Image{EnvVariables: map[string]string{"foo": "bar"}}},
			want:   map[string]string{"foo": "bar", "test": "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := Image{
				Parent:       tt.fields.Parent,
				EnvVariables: tt.fields.EnvVariables,
			}
			assert.Equalf(t, tt.want, im.GetAllEnvVar(), "GetAllEnvVar()")
		})
	}
}

func TestImage_GetAllPackages(t *testing.T) {
	type fields struct {
		Parent   *Image
		Packages map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "SuccessNoPackage",
			want: map[string]string{},
		},
		{
			name:   "SuccessPackageWithNoParent",
			fields: fields{Packages: map[string]string{"test": "test"}},
			want:   map[string]string{"test": "test"},
		},
		{
			name:   "SuccessPackageWithParent",
			fields: fields{Packages: map[string]string{"test": "test"}, Parent: &Image{Packages: map[string]string{"foo": "bar"}}},
			want:   map[string]string{"foo": "bar", "test": "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := Image{
				Parent:   tt.fields.Parent,
				Packages: tt.fields.Packages,
			}
			assert.Equalf(t, tt.want, im.GetAllPackages(), "GetAllPackages()")
		})
	}
}

func TestImage_GetTags(t *testing.T) {
	type fields struct {
		Name  string
		Tag   string
		Alias []ImageName
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Success",
			fields: fields{
				Name: "test",
				Tag:  "0.1",
			},
			want: []string{"test:0.1"},
		},
		{
			name: "SuccessWithAlias",
			fields: fields{
				Name:  "test",
				Tag:   "0.1",
				Alias: []ImageName{{Name: "foo", Tag: "0.2"}},
			},
			want: []string{"test:0.1", "foo:0.2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := Image{
				ImageName: ImageName{
					Name: tt.fields.Name,
					Tag:  tt.fields.Tag,
				},
				Alias: tt.fields.Alias,
			}
			assert.Equalf(t, tt.want, im.GetNames(), "GetTags()")
		})
	}
}
