package template

import (
	"fmt"
	"github.com/alexandreh2ag/mib/assets"
	"github.com/alexandreh2ag/mib/context"
	"github.com/alexandreh2ag/mib/types"
	"github.com/alexandreh2ag/mib/version"
	"github.com/spf13/afero"
	"path/filepath"
	"text/template"
	"time"
)

var (
	IndexTmplPath = "tmpl/index-readme.tmpl"
	ImageTmplPath = "tmpl/image-readme.tmpl"
)

func GenerateTemplate(ctx *context.Context, tmplPath string, data any, outputPath string) error {
	content, err := assets.GetTmplContent(tmplPath)
	if err != nil {
		return err
	}
	additionalVars := template.FuncMap{
		"now":     time.Now,
		"getUrl":  types.GetUrl,
		"config":  ctx.Config.Get,
		"version": version.GetFormattedVersion,
	}

	tmpl, err := template.New(tmplPath).Funcs(additionalVars).Parse(content)
	if err != nil {
		return err
	}

	file, err := ctx.FS.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

func GenerateReadmeIndex(ctx *context.Context, images types.Images, outputPath string) error {
	return GenerateTemplate(ctx, IndexTmplPath, images, outputPath)
}

func GenerateReadmeImages(ctx *context.Context, images types.Images) error {
	for _, image := range images {
		err := GenerateTemplate(ctx, ImageTmplPath, image, fmt.Sprintf("%s/README.md", image.Path))
		if err != nil {
			return err
		}
	}
	return nil
}

func OverrideTemplatesFromConfig(ctx *context.Context) error {
	templateCfg := ctx.Config.Template
	if templateCfg.IndexPath != "" {
		content, err := GetTemplateFileContent(ctx, templateCfg.IndexPath)
		if err != nil {
			return err
		}
		_ = assets.SeTmplContent(IndexTmplPath, string(content))
	}

	if templateCfg.ImagePath != "" {
		content, err := GetTemplateFileContent(ctx, templateCfg.ImagePath)
		if err != nil {
			return err
		}
		_ = assets.SeTmplContent(ImageTmplPath, string(content))

	}

	return nil
}

func GetTemplateFileContent(ctx *context.Context, path string) ([]byte, error) {
	afs := &afero.Afero{Fs: ctx.FS}
	if !filepath.IsAbs(path) {
		path = filepath.Join(ctx.WorkingDir, path)
	}
	return afs.ReadFile(path)
}
