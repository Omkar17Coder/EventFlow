package config

type ChannelsConfig struct {
	API   APIConfig   `yaml:"api"`
	DB    DBConfig    `yaml:"db"`
	Email EmailConfig `yaml:"email"`
}

func DefaultChannelsConfig() ChannelsConfig {
	return ChannelsConfig{
		API:   DefaultAPIConfig(),
		DB:    DefaultDBConfig(),
		Email: DefaultEmailConfig(),
	}
}
