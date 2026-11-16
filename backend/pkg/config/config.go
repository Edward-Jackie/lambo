package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func NewConfig() *viper.Viper {
	envConf := os.Getenv("APP_CONF")
	if envConf == "" {
		flag.StringVar(&envConf, "conf", "config/local.yml", "config path, eg: -conf config/local.yml")
		flag.Parse()
	}
	if envConf == "" {
		envConf = "config/local.yml"
	}
	fmt.Println("load conf file:", envConf)
	return getConfig(envConf)

}
func getConfig(path string) *viper.Viper {
	conf := viper.New()
	for key, env := range map[string]string{
		"data.postgres.dsn":         "LAMBO_POSTGRES_DSN",
		"data.auto_migrate":         "LAMBO_AUTO_MIGRATE",
		"bootstrap.enabled":         "LAMBO_BOOTSTRAP_ENABLED",
		"bootstrap.admin_username":  "LAMBO_ADMIN_USERNAME",
		"bootstrap.admin_password":  "LAMBO_ADMIN_PASSWORD",
		"bootstrap.reader_username": "LAMBO_READER_USERNAME",
		"bootstrap.reader_password": "LAMBO_READER_PASSWORD",
	} {
		_ = conf.BindEnv(key, env)
	}
	conf.SetConfigFile(path)
	err := conf.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return conf
}
