package config

type Config struct {
	Build    Build    `mapstructure:"build"`
	Template Template `mapstructure:"template"`
}
type Build struct {
	ExtensionExclude string `mapstructure:"extensionExclude"`
}

type Template struct {
	ImagePath string `mapstructure:"imagePath"`
	IndexPath string `mapstructure:"indexPath"`
}

func NewConfig() Config {
	return Config{}
}

func DefaultConfig() Config {
	cfg := NewConfig()

	cfg.Build.ExtensionExclude = ".md,.txt"

	return cfg
}
