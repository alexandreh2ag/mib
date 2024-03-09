package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfig(t *testing.T) {
	got := NewConfig()

	assert.Equal(t, Config{}, got)
}

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()
	want := Config{
		Build: Build{
			ExtensionExclude: ".md,.txt",
		},
	}
	assert.Equal(t, want, got)
}
