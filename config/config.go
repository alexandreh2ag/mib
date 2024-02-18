package config

type Config struct {
	Build struct {
		ExtensionExclude string `mapstructure:"extensionExclude"`
	} `mapstructure:"build"`
	Template struct {
		ImagePath string `mapstructure:"imagePath"`
		IndexPath string `mapstructure:"indexPath"`
	} `mapstructure:"template"`
}

func (c Config) Get() Config {
	return c
}

func NewConfig() Config {
	return Config{}
}

func DefaultConfig() Config {
	cfg := NewConfig()

	cfg.Build.ExtensionExclude = ".md,.txt"

	return cfg
}
