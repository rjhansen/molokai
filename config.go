package main

import (
	"github.com/spf13/viper"
)

func loadConfig() {
	viper.SetDefault("host", "localhost")
	viper.SetDefault("user", "molokai")
	viper.SetDefault("database", "sensor_readings")
	viper.SetConfigFile("/etc/molokai.conf")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic("configuration file error, bailing out")
	}
}
