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
func LoadConfig(configFile string) (config *Config, err error) {
	log.Debug().Msgf("config:: reading config from: %s", configFile)
	viper.SetConfigFile(configFile)
	err = viper.ReadInConfig()
	if err != nil {
		commonutils.ReportOnError(err, "config:: read error")
		return nil, err
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		commonutils.ReportOnError(err, "config:: unmarshal error")
		return nil, err
	}
	err = loadReceiver(config)
	if err != nil {
		return nil, err
	}
	err = loadPublisher(config)
	if err != nil {
		return nil, err
	}
	err = viper.UnmarshalKey("log", &config.LogConfig)
	if err != nil {
		commonutils.ReportOnError(err, "config:: log unmarshal error")
		return nil, err
	}
	return config, err
}

func loadReceiver(config *Config) error {
	var err error
	switch config.ReceiverType {
	case "rmq":
		err = viper.UnmarshalKey("rec_rmq", &config.ReceiveRmq)
		if err != nil {
			commonutils.ReportOnError(err, "config:: rec_rmq config unmarshal error")
			return err
		}
		return nil
	default:
		commonutils.ReportOnError(errors.New("no_receiver_type"), "config::")
		return errors.New("no receiver config")
	}
}
func loadPublisher(config *Config) error {
	var err error
	switch config.PublisherType {
	case "rmq":
		err = viper.UnmarshalKey("pub_rmq", &config.PublishRmq)
		if err != nil {
			commonutils.ReportOnError(err, "config:: pub_rmq config unmarshal error")
			return err
		}
		return nil
	default:
		commonutils.ReportOnError(errors.New("no_publisher_type"), "config::")
		return errors.New("no publisher config")
	}
}
