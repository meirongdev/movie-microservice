package config

import "fmt"

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (c MySQLConfig) FormatDSN() string {
	// Format DSN string
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", c.Username, c.Password, c.Host, c.Database)
}
