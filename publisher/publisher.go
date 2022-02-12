package publisher

import (
	"errors"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
)

func ConfigurePublisher(cfg *config.Config) (chan<- *models.Result, <-chan error, error) {
	if cfg.PublishRmq != nil {
		return initRmqPublisher(cfg.PublishRmq)
	}
	return nil, nil, errors.New("no_publisher_configured")
}

func ClosePublisher(_ *config.Config) {
	//
}
