package loader

import (
	"github.com/alexandreh2ag/mib/context"
	mock_validator "github.com/alexandreh2ag/mib/mock/validator"
	"github.com/alexandreh2ag/mib/types"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

func TestLoadImage(t *testing.T) {
	type args struct {
		ctx  *context.Context
		path string
	}
	tests := []struct {
		name    string
		args    args
		preRun  func(ctx *context.Context)
		want    types.Image
		wantErr bool
	}{
		{
			name: "CheckOk",
			args: args{ctx: context.TestContext(nil), path: "/app/test/mib.yml"},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian:latest"), 0644)
			},
			want: types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}},
		},
		{
			name:    "CheckFailedWhenDBPFileNotExist",
			args:    args{ctx: context.TestContext(nil), path: "/app/test/mib.yml"},
			preRun:  func(ctx *context.Context) {},
			want:    types.Image{Path: "/app/test", RelativeDir: "test"},
			wantErr: true,
		},
		{
			name: "CheckFailedWhenDBPFileBadFormat",
			args: args{ctx: context.TestContext(nil), path: "/app/test/mib.yml"},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: {test"), 0644)
			},
			want:    types.Image{Path: "/app/test", RelativeDir: "test"},
			wantErr: true,
		},
		{
			name: "CheckFailedWhenDockerfileNotExist",
			args: args{ctx: context.TestContext(nil), path: "/app/test/mib.yml"},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1"), 0644)
			},
			want:    types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preRun(tt.args.ctx)
			got, err := LoadImage(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findParentImage(t *testing.T) {
	type args struct {
		ctx   *context.Context
		image *types.Image
	}
	tests := []struct {
		name    string
		args    args
		preRun  func(ctx *context.Context)
		want    *types.Image
		wantErr bool
	}{
		{
			name: "CheckOK",
			args: args{ctx: context.TestContext(nil), image: &types.Image{Path: "/app/test"}},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian:latest"), 0644)
			},
			want: &types.Image{Path: "/app/test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}},
		},
		{
			name:    "CheckFailedWhenDockerfileNotExist",
			args:    args{ctx: context.TestContext(nil), image: &types.Image{Path: "/app/test"}},
			preRun:  func(ctx *context.Context) {},
			want:    &types.Image{Path: "/app/test"},
			wantErr: true,
		},
		{
			name: "CheckFailedWhenParentNameIsNotParsable",
			args: args{ctx: context.TestContext(nil), image: &types.Image{Path: "/app/test"}},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian"), 0644)
			},
			want:    &types.Image{Path: "/app/test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preRun(tt.args.ctx)
			if err := findParentImage(tt.args.ctx, tt.args.image); (err != nil) != tt.wantErr {
				t.Errorf("findParentImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.image, tt.want) {
				t.Errorf("findParentImage() got = %v, want %v", tt.args.image, tt.want)
			}
		})
	}
}

func TestLoadImages(t *testing.T) {
	type args struct {
		ctx *context.Context
	}
	tests := []struct {
		name    string
		args    args
		preRun  func(ctx *context.Context)
		want    types.Images
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "CheckOKWhenEmpty",
			args:    args{ctx: context.TestContext(nil)},
			preRun:  func(ctx *context.Context) {},
			want:    types.Images{},
			wantErr: assert.NoError,
		},
		{
			name: "CheckOKWhenNoFileMatch",
			args: args{ctx: context.TestContext(nil)},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/test", []byte("content"), 0644)
			},
			want:    types.Images{},
			wantErr: assert.NoError,
		},
		{
			name: "CheckOKWithOneImage",
			args: args{ctx: context.TestContext(nil)},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian:latest"), 0644)
			},
			want: types.Images{
				&types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}},
			},
			wantErr: assert.NoError,
		},
		{
			name: "CheckOKWithTwoImage",
			args: args{ctx: context.TestContext(nil)},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian:latest"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.3"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:dev"), 0644)
			},
			want: types.Images{
				&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.3"}, Path: "/app/foo", RelativeDir: "foo", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "dev"}}},
				&types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}},
			},
			wantErr: assert.NoError,
		},
		{
			name: "SuccessWithChild",
			args: args{ctx: context.TestContext(nil)},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian:latest"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM test:0.1"), 0644)
			},
			want: func() types.Images {
				parent := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}}
				child := &types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}, HasLocalParent: true, HasParentToBuild: true, Path: "/app/foo", RelativeDir: "foo", Parent: parent}
				parent.Children = types.Images{child}
				return types.Images{
					parent,
				}
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "CheckOkWithOneImageFail",
			args: args{ctx: context.TestContext(nil)},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.3"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM debian:dev"), 0644)
			},
			want: types.Images{
				&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.3"}, Path: "/app/foo", RelativeDir: "foo", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "dev"}}},
			},
			wantErr: assert.NoError,
		},
		{
			name: "FailValidator",
			args: args{ctx: context.TestContext(nil)},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1\nalias: [{name: ''}]"), 0644)
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian:latest"), 0644)
			},
			want: types.Images{
				&types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}, Alias: []types.ImageName{{}}},
			},
			wantErr: assert.Error,
		},
		{
			name: "FailChildValidator",
			args: args{ctx: context.TestContext(nil)},
			preRun: func(ctx *context.Context) {
				afero.WriteFile(ctx.FS, "/app/test/mib.yml", []byte("name: test\ntag: 0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/test/Dockerfile", []byte("FROM debian:latest"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/mib.yml", []byte("name: foo\ntag: 0.1\nalias: [{name: ''}]"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo/Dockerfile", []byte("FROM test:0.1"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo2/mib.yml", []byte("name: foo2\ntag: 0.1\nalias: [{name: ''}]"), 0644)
				afero.WriteFile(ctx.FS, "/app/foo2/Dockerfile", []byte("FROM foo:0.1"), 0644)
			},
			want: func() types.Images {
				parent := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}}
				child := &types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}, Alias: []types.ImageName{{}}, HasLocalParent: true, HasParentToBuild: true, Path: "/app/foo", RelativeDir: "foo", Parent: parent}
				childChild := &types.Image{ImageName: types.ImageName{Name: "foo2", Tag: "0.1"}, Alias: []types.ImageName{{}}, HasLocalParent: true, HasParentToBuild: true, Path: "/app/foo2", RelativeDir: "foo2", Parent: child}
				parent.Children = types.Images{child}
				child.Children = types.Images{childChild}
				return types.Images{
					parent,
				}
			}(),
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preRun(tt.args.ctx)
			got, err := LoadImages(tt.args.ctx)
			tt.wantErr(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadImages() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_orderDependencyImages(t *testing.T) {

	image1 := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}}
	image2 := &types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.3"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "dev"}}}

	imageChild := &types.Image{ImageName: types.ImageName{Name: "bar", Tag: "1.0"}, Parent: &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}}}
	imageChildChild := &types.Image{ImageName: types.ImageName{Name: "bar2", Tag: "1.1"}, Parent: &types.Image{ImageName: types.ImageName{Name: "bar", Tag: "1.0"}}}

	imagePlatform := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Platforms: []string{"foo", "bar"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}}
	imageChildPlatform := &types.Image{ImageName: types.ImageName{Name: "bar", Tag: "1.0"}, Parent: &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}}}
	tests := []struct {
		name         string
		imagesToSort types.Images
		want         types.Images
	}{
		{
			name:         "CheckOkWithNoImage",
			imagesToSort: types.Images{},
			want:         types.Images{},
		},
		{
			name: "CheckOkWithOneImage",
			imagesToSort: types.Images{
				image1,
				image2,
			},
			want: types.Images{
				&types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}},
				&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.3"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "dev"}}},
			},
		},
		{
			name: "CheckOkWithTwoImage",
			imagesToSort: types.Images{
				&types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}},
				&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.3"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "dev"}}},
			},
			want: types.Images{
				&types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}},
				&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.3"}, Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "dev"}}},
			},
		},
		{
			name: "CheckOkWithTwoImageAndOneChild",
			imagesToSort: types.Images{
				image1,
				image2,
				imageChild,
			},
			want: types.Images{
				image1,
				image2,
			},
		},
		{
			name: "CheckOkWithTwoImageAndOneChildOfChild",
			imagesToSort: types.Images{
				image1,
				image2,
				imageChild,
				imageChildChild,
			},
			want: types.Images{
				image1,
				image2,
			},
		},
		{
			name: "CheckOkWithPlatform",
			imagesToSort: types.Images{
				imagePlatform,
				imageChildPlatform,
			},
			want: types.Images{
				imagePlatform,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orderDependencyImages(tt.imagesToSort); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("orderDependencyImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveExtExcludePath(t *testing.T) {
	type args struct {
		workingDir       string
		extensionExclude string
		filesUpdated     []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "SuccessNothing",
			args: args{
				workingDir:       "/app",
				extensionExclude: "",
				filesUpdated:     []string{},
			},
			want: []string{},
		},
		{
			name: "SuccessWithoutExclude",
			args: args{
				workingDir:       "/app",
				extensionExclude: "",
				filesUpdated:     []string{"foo/Dockerfile", "bar/Dockerfile"},
			},
			want: []string{"/app/foo/Dockerfile", "/app/bar/Dockerfile"},
		},
		{
			name: "SuccessWithExclude",
			args: args{
				workingDir:       "/app",
				extensionExclude: ".md,.txt",
				filesUpdated:     []string{"foo/file.md", "foo/file.txt", "bar/Dockerfile"},
			},
			want: []string{"/app/bar/Dockerfile"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveExtExcludePath(tt.args.workingDir, tt.args.extensionExclude, tt.args.filesUpdated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveExtExcludePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reformatValidatorError(t *testing.T) {

	tests := []struct {
		name       string
		imagesFunc func() types.Images
		errFunc    func(*gomock.Controller) validator.FieldError
		want       string
	}{
		{
			name: "SuccessReplaceIndex",
			imagesFunc: func() types.Images {
				return types.Images{&types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}}}
			},
			errFunc: func(ctrl *gomock.Controller) validator.FieldError {
				errMock := mock_validator.NewMockFieldError(ctrl)
				errMock.EXPECT().Namespace().Times(1).Return("[0].Alias[0].Name")
				errMock.EXPECT().Error().Times(1).Return("Key: '[0].Alias[0].Name' Error:Field validation for 'Name' failed on the 'required' tag")
				return errMock
			},
			want: "Key: '/app/test Alias[0].Name' Error:Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name: "SuccessReplaceIndexChild",
			imagesFunc: func() types.Images {
				parent := &types.Image{ImageName: types.ImageName{Name: "test", Tag: "0.1"}, Path: "/app/test", RelativeDir: "test", Parent: &types.Image{ImageName: types.ImageName{Name: "debian", Tag: "latest"}}}
				child := &types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}, Alias: []types.ImageName{{}}, HasLocalParent: true, HasParentToBuild: true, Path: "/app/foo", RelativeDir: "foo", Parent: parent}
				childChild := &types.Image{ImageName: types.ImageName{Name: "foo2", Tag: "0.1"}, Alias: []types.ImageName{{}}, HasLocalParent: true, HasParentToBuild: true, Path: "/app/foo2", RelativeDir: "foo2", Parent: child}
				parent.Children = types.Images{child}
				child.Children = types.Images{childChild}
				return types.Images{
					parent,
				}
			},
			errFunc: func(ctrl *gomock.Controller) validator.FieldError {
				errMock := mock_validator.NewMockFieldError(ctrl)
				errMock.EXPECT().Namespace().Times(1).Return("[0].Children[0].Children[0].Alias[0].Name")
				errMock.EXPECT().Error().Times(1).Return("Key: '[0].Children[0].Children[0].Alias[0].Name' Error:Field validation for 'Name' failed on the 'required' tag")
				return errMock
			},
			want: "Key: '/app/foo2 Alias[0].Name' Error:Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name: "SuccessDefaultError",
			imagesFunc: func() types.Images {
				return types.Images{}
			},
			errFunc: func(ctrl *gomock.Controller) validator.FieldError {
				errMock := mock_validator.NewMockFieldError(ctrl)
				errMock.EXPECT().Namespace().Times(1).Return("not matching")
				errMock.EXPECT().Error().Times(1).Return("Key: Error:Field validation for 'Name' failed on the 'required' tag")
				return errMock
			},
			want: "Key: Error:Field validation for 'Name' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			err := tt.errFunc(ctrl)
			got := reformatValidatorError(tt.imagesFunc(), err)
			assert.Equal(t, tt.want, got)
		})
	}
}
