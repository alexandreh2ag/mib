package assets

import (
	"embed"
	"fmt"
	"io/fs"
)

const (
	templatesDir = "data"
)

var (
	//go:embed data/*
	files     embed.FS
	templates = map[string]string{}
)

func init() {
	LoadTemplates(templates, templatesDir)
}

func LoadTemplates(dest map[string]string, pathRead string) {

	paths, err := fs.ReadDir(files, pathRead)
	if err != nil {
		panic(err)
	}

	for _, path := range paths {
		fullPath := fmt.Sprintf("%s/%s", pathRead, path.Name())
		if path.IsDir() {
			LoadTemplates(dest, fullPath)
			continue
		}
		content, errRead := fs.ReadFile(files, fullPath)
		if errRead != nil {
			panic(errRead)
		}
		dest[fullPath] = string(content)
	}
}

func GetTmplContent(filePath string) (string, error) {
	fullPath := fmt.Sprintf("%s/%s", templatesDir, filePath)
	if tmpl, ok := templates[fullPath]; ok {
		return tmpl, nil
	}
	return "", fmt.Errorf("template %s not found", filePath)
}

func SeTmplContent(filePath string, content string) error {
	fullPath := fmt.Sprintf("%s/%s", templatesDir, filePath)

	templates[fullPath] = content

	return nil
}

func GetEmbedFiles() *embed.FS {
	return &files
}
