package config

type Config struct {
	Build    Build    `mapstructure:"build"`
	Template Template `mapstructure:"template" validate:"omitempty,required"`
}

func (g *Config) Get() *Config {
	return g
}

type Build struct {
	ExtensionExclude string `mapstructure:"extensionExclude" validate:"required"`
}

type Template struct {
	ImagePath string `mapstructure:"imagePath" validate:"omitempty,required"`
	IndexPath string `mapstructure:"indexPath" validate:"omitempty,required"`
}

func NewConfig() Config {
	return Config{}
}

func DefaultConfig() Config {
	cfg := NewConfig()

	cfg.Build.ExtensionExclude = ".md,.txt"

	return cfg
}
