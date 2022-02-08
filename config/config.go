package config

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
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

type LogConfig struct {
	Agent string `mapstructure:"agent"`
	Level string `mapstructure:"level"`
	// file logger
	FilePath string `mapstructure:"file_path"`
}

type Config struct {
	Concurrency  int             `mapstructure:"concurrency"`
	ReceiverType string          `mapstructure:"receiver_type"`
	ReceiveRmq   *RabbitmqConfig `mapstructure:"rec_rmq"`

	PublisherType string          `mapstructure:"publisher_type"`
	PublishRmq    *RabbitmqConfig `mapstructure:"pub_rmq"`

	LogConfig *LogConfig `mapstructure:"log"`
}

func RmqUrl(rmqConfig *RabbitmqConfig) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", rmqConfig.Username, rmqConfig.Password, rmqConfig.Host, rmqConfig.Port)
}

//LoadConfig This method does not return the error, because we will not boot up for config related error
func LoadConfig(configFile string) (config *Config) {
	log.Debug().Msgf("config:: reading config from: %s", configFile)
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	commonutils.ExitOnError(err, "config:: read error")
	err = viper.Unmarshal(&config)
	commonutils.ExitOnError(err, "config:: unmarshal error")
	loadReceiver(config)
	loadPublisher(config)
	err = viper.UnmarshalKey("log", &config.LogConfig)
	commonutils.ExitOnError(err, "config:: log unmarshal error")
	return config
}

func loadReceiver(config *Config) {
	var err error
	switch config.ReceiverType {
	case "rmq":
		err = viper.UnmarshalKey("rec_rmq", &config.ReceiveRmq)
		commonutils.ExitOnError(err, "config:: rec_rmq config unmarshal error")
		break
	default:
		commonutils.ExitOnError(errors.New("no_receiver_type"), "Config::")
	}
}
func loadPublisher(config *Config) {
	var err error
	switch config.PublisherType {
	case "rmq":
		err = viper.UnmarshalKey("rec_rmq", &config.PublishRmq)
		commonutils.ExitOnError(err, "config:: pub_rmq config unmarshal error")
		break
	default:
		commonutils.ExitOnError(errors.New("no_publisher_type"), "Config::")
	}
}
