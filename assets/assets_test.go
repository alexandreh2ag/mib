package assets

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSeTmplContent_Success(t *testing.T) {
	filePath := "test"
	content := "test content"
	err := SeTmplContent(filePath, content)
	assert.NoErrorf(t, err, "SeTmplContent() error = %v", err)
	got := templates[fmt.Sprintf("%s/%s", templatesDir, filePath)]
	assert.Equal(t, content, got)
}

func TestGetTmplContent_Success(t *testing.T) {
	filePath := "test_Success"
	content := "test content"
	templates[fmt.Sprintf("%s/%s", templatesDir, filePath)] = content
	got, err := GetTmplContent(filePath)
	assert.NoErrorf(t, err, "SeTmplContent() error = %v", err)
	assert.Equalf(t, content, got, "GetTmplContent(%v)", filePath)
}

func TestGetTmplContent_Fail(t *testing.T) {
	filePath := "test_fail"
	got, err := GetTmplContent(filePath)
	assert.Errorf(t, err, "SeTmplContent() error = %v", err)
	assert.Equalf(t, "", got, "GetTmplContent(%v)", filePath)
}

func TestLoadTemplates_Success(t *testing.T) {
	destTmpl := map[string]string{}
	LoadTemplates(destTmpl, templatesDir)
	assert.Equalf(t, 2, len(destTmpl), "destTmpl count miss match")
}

func TestLoadTemplates_Fail(t *testing.T) {
	destTmpl := map[string]string{}
	defer func() {
		if r := recover(); r != nil {
			assert.True(t, true)
		} else {
			t.Errorf("LoadTemplates should have panicked")
		}
	}()
	LoadTemplates(destTmpl, "/wrong")
}

func TestGetEmbedFiles(t *testing.T) {
	got := GetEmbedFiles()
	assert.NotNil(t, got)
}
