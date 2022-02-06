package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
)

type RabbitmqConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Queue    string `mapstructure:"queue"`
}

type Config struct {
	Concurrency  int    `mapstructure:"concurrency"`
	ReceiverType string `mapstructure:"receiver_type"`
	Rabbitmq     *RabbitmqConfig
}

func RmqUrl(cfg *Config) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.Rabbitmq.Username, cfg.Rabbitmq.Password, cfg.Rabbitmq.Host, cfg.Rabbitmq.Port)
}

//LoadConfig This method does not return the error, because we will not boot up for config related error
func LoadConfig(configFile string) (config *Config) {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	commonutils.ExitOnError(err, "Config read error")
	err = viper.Unmarshal(&config)
	commonutils.ExitOnError(err, "Config unmarshal error")
	switch config.ReceiverType {
	case "rmq":
		err = viper.UnmarshalKey("rmq", &config.Rabbitmq)
		break
	//case "cli":
	//	err = viper.UnmarshalKey("cli", &config.Cli)
	//	break
	default:
		commonutils.ExitOnError(errors.New("No receiver_type"), "Rabbitmq config unmarshal error")
	}
	return config
}
