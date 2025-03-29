package config

type DatabaseConfig struct {
	Driver                  string
	Url                     string
	IsSQLite                bool
	ConnMaxLifetimeInMinute int
	MaxOpenConns            int
	MaxIdleConns            int
}
