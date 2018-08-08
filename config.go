package monger

import (
	"gopkg.in/mgo.v2"
)

type Config struct {
	Hosts     []string
	DBName    string
	User      string
	Password  string
	PoolLimit int
	DialInfo  *mgo.DialInfo
}

type ConfigOption func(*Config)

func PoolLimit(poolLimit int) ConfigOption {
	return func(c *Config) {
		c.PoolLimit = poolLimit
	}
}

func DBName(name string) ConfigOption {
	return func(c *Config) {
		c.DBName = name
	}
}

func Hosts(hosts []string) ConfigOption {
	return func(c *Config) {
		c.Hosts = hosts
	}
}

func User(user string) ConfigOption {
	return func(c *Config) {
		c.User = user
	}
}

func Password(password string) ConfigOption {
	return func(c *Config) {
		c.Password = password
	}
}

func DialInfo(info *mgo.DialInfo) ConfigOption {
	return func(c *Config) {
		c.DialInfo = info
	}
}
