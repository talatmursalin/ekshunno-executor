package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"strings"
)

type RabbitmqConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Queue    string `mapstructure:"queue"`
}

type PublishFileConfig struct {
	Name string `mapstructure:"name"`
}

type Config struct {
	Concurrency  int             `mapstructure:"concurrency"`
	ReceiverType string          `mapstructure:"receiver_type"`
	ReceiveRmq   *RabbitmqConfig `mapstructure:"rec_rmq"`

	PublisherType string             `mapstructure:"publisher_type"`
	PublishRmq    *RabbitmqConfig    `mapstructure:"pub_rmq"`
	PublishFile   *PublishFileConfig `mapstructure:"pub_file"`
}

func RmqUrl(rmqConfig *RabbitmqConfig) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", rmqConfig.Username, rmqConfig.Password, rmqConfig.Host, rmqConfig.Port)
}

//LoadConfig This method does not return the error, because we will not boot up for config related error
func LoadConfig(configFile string) (config *Config) {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	commonutils.ExitOnError(err, "Config read error")
	err = viper.Unmarshal(&config)
	commonutils.ExitOnError(err, "Config unmarshal error")
	loadReceiver(config)
	loadPublisher(config)
	return config
}

func loadReceiver(config *Config) {
	var err error
	switch config.ReceiverType {
	case "rmq":
		err = viper.UnmarshalKey("rec_rmq", &config.ReceiveRmq)
		commonutils.ExitOnError(err, "rec_rmq config unmarshal error")
		break
	default:
		commonutils.ExitOnError(errors.New("No receiver_type"), "")
	}
}
func loadPublisher(config *Config) {
	var err error
	all_publishers := strings.Split(config.PublisherType, "|")
	for _, pub := range all_publishers {
		switch pub {
		case "pub_rmq":
			err = viper.UnmarshalKey("pub_rmq", &config.PublishRmq)
			commonutils.ExitOnError(err, "pub_rmq config unmarshal error")
			break
		case "file":
			err = viper.UnmarshalKey("pub_file", &config.PublishFile)
			commonutils.ExitOnError(err, "pub_file config unmarshal error")
		}
	}
}
