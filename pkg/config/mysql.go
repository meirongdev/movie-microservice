package config

import "fmt"

type MySQLConfig struct {
	Host     string
	Username string
	Password string
	Database string
}

func (c MySQLConfig) FormatDSN() string {
	// Format DSN string
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", c.Username, c.Password, c.Host, c.Database)
}
