package template

import (
	"fmt"
	"github.com/alexandreh2ag/mib/assets"
	"github.com/alexandreh2ag/mib/config"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGenerateTemplate(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx := &context.Context{
		FS:     afero.NewMemMapFs(),
		Config: &cfg,
	}
	ctx.WorkingDir = "/test"
	outputPath := fmt.Sprintf("%s/README.md", ctx.WorkingDir)
	_ = ctx.FS.MkdirAll(ctx.WorkingDir, os.FileMode(0664))
	afs := &afero.Afero{Fs: ctx.FS}
	tmplNameGood := "GenerateTemplate_Good"
	tmplNameWrong := "GenerateTemplate_Wrong"
	_ = assets.SeTmplContent(tmplNameGood, "{{ .Name }}")
	_ = assets.SeTmplContent(tmplNameWrong, "{{ .Test }}")

	tests := []struct {
		name       string
		data       types.ImageName
		outputPath string
		tmplPath   string
		checkFn    func(t *testing.T, data string)
		want       string
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "Success",
			data:       types.ImageName{Name: "foo"},
			tmplPath:   tmplNameGood,
			want:       "foo",
			outputPath: outputPath,
			wantErr:    assert.NoError,
			checkFn: func(t *testing.T, data string) {
				content, err := afs.ReadFile(outputPath)
				assert.Nil(t, err)
				assert.Equal(t, data, string(content))
			},
		},
		{
			name:     "Fail_WhenTmplPathDoesNotExist",
			data:     types.ImageName{Name: "foo"},
			tmplPath: "path-not-exist",
			wantErr:  assert.Error,
		},
		{
			name:     "Fail_WhenTmplCannotExec",
			data:     types.ImageName{Name: "foo"},
			tmplPath: tmplNameWrong,
			wantErr:  assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateTemplate(ctx, tt.tmplPath, tt.data, tt.outputPath)

			if err != nil {
				tt.wantErr(t, err, "GenerateTemplate is not valid")
			}
			if tt.checkFn != nil {
				tt.checkFn(t, tt.want)
			}
		})
	}
}

func TestGenerateReadmeIndex(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx := &context.Context{
		FS:     afero.NewMemMapFs(),
		Config: &cfg,
	}
	ctx.WorkingDir = "/test"
	_ = ctx.FS.MkdirAll(ctx.WorkingDir, os.FileMode(0664))
	images := types.Images{}
	err := GenerateReadmeIndex(ctx, images, fmt.Sprintf("%s/README.md", ctx.WorkingDir))
	assert.NoError(t, err)
}

func TestGetTemplateFileContent(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx := &context.Context{
		FS:     afero.NewMemMapFs(),
		Config: &cfg,
	}
	ctx.WorkingDir = "test"
	afs := &afero.Afero{Fs: ctx.FS}
	_ = afs.WriteFile("test/assert.tmpl", []byte("test"), os.FileMode(0664))
	got, err := GetTemplateFileContent(ctx, "assert.tmpl")
	assert.Nil(t, err)
	assert.Equal(t, []byte("test"), got)
}

func TestOverrideTemplatesFromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx := &context.Context{
		FS:         afero.NewMemMapFs(),
		Config:     &cfg,
		WorkingDir: "/app",
	}
	indexTmplPath := fmt.Sprintf("%s/index.tmpl", ctx.WorkingDir)
	imageTmplPath := fmt.Sprintf("%s/image.tmpl", ctx.WorkingDir)

	afs := &afero.Afero{Fs: ctx.FS}
	_ = afs.WriteFile(indexTmplPath, []byte("test"), os.FileMode(0664))
	_ = afs.WriteFile(imageTmplPath, []byte("test"), os.FileMode(0664))
	tests := []struct {
		name    string
		tmplCfg config.Template
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "SuccessNothing",
			wantErr: assert.NoError,
		},
		{
			name:    "SuccessOverrideIndex",
			tmplCfg: config.Template{IndexPath: indexTmplPath},
			wantErr: assert.NoError,
		},
		{
			name:    "FailOverrideIndex",
			tmplCfg: config.Template{IndexPath: "/no-exist"},
			wantErr: assert.Error,
		},
		{
			name:    "SuccessOverrideImage",
			tmplCfg: config.Template{ImagePath: imageTmplPath},
			wantErr: assert.NoError,
		},
		{
			name:    "FailOverrideImage",
			tmplCfg: config.Template{ImagePath: "/no-exist"},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.Config.Template = tt.tmplCfg
			tt.wantErr(t, OverrideTemplatesFromConfig(ctx), fmt.Sprintf("OverrideTemplatesFromConfig(%v)", ctx))
		})
	}
}

func TestGenerateReadmeImages_EmptyImages(t *testing.T) {
	ctx := context.TestContext(nil)
	afs := &afero.Afero{Fs: ctx.FS}
	path := ctx.WorkingDir
	_ = afs.Mkdir(path, 0775)
	images := types.Images{}
	err := GenerateReadmeImages(ctx, images)
	assert.NoError(t, err)
}

func TestGenerateReadmeImages_WithImages(t *testing.T) {
	ctx := context.TestContext(nil)
	afs := &afero.Afero{Fs: ctx.FS}
	path := ctx.WorkingDir
	_ = afs.Mkdir(path, 0775)
	images := types.Images{&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}, Path: fmt.Sprintf("%s/foo", path)}}
	err := GenerateReadmeImages(ctx, images)
	assert.NoError(t, err)
	exist, err := afs.Exists(fmt.Sprintf("%s/foo/README.md", path))
	assert.NoError(t, err)
	assert.True(t, exist)
}

func TestGenerateReadmeImages_Error(t *testing.T) {
	ctx := context.TestContext(nil)
	afs := &afero.Afero{Fs: ctx.FS}
	path := ctx.WorkingDir
	_ = afs.Mkdir(path, 0775)
	ImageTmplPath = "wrong-path.tmpl"
	images := types.Images{&types.Image{ImageName: types.ImageName{Name: "foo", Tag: "0.1"}, Path: fmt.Sprintf("%s/foo", path)}}
	err := GenerateReadmeImages(ctx, images)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template wrong-path.tmpl not found")
}
