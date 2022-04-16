package config

import (
	"fmt"
	"zoo/util"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var Config = struct {
	AppName    string
	HTTPServer struct {
		Address string
	}
	TcpServer struct {
		Address string
	}
	Redis struct {
		Address  string
		Password string
		Database int
	}
	Mysql struct {
		Address  string
		Username string
		Password string
		Database string
	}
}{}

func init() {
	// live deploy:
	// 1. edit config file: /opt/data/config/config.live.yml
	// 2. start app: ZOO_PROFILE=live ./zoo
	profile := strings.ToLower(util.Env.GetEnvDefault("ZOO_PROFILE", "dev"))
	configName := fmt.Sprintf("config.%s.yml", profile)
	log.Printf("load config from %s\n", configName)
	viper.AutomaticEnv()
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("/opt/data/config/") // 线上服务器，会将配置放此路径下
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("read config failed: %v", err)
	}
	viper.Unmarshal(&Config)
}
