package logwriter

type Config struct {
	User     string `yaml:"User" validate:"nonzero"`
	Port     int    `yaml:"Port" validate:"nonzero"`
	Password string `yaml:"Password" validate:"nonzero"`
	LogPath  string `yaml:"LogPath" validate:"nonzero"`
	Interval int    `yaml:"Interval" validate:"nonzero"`
}
