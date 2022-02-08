package receiver

import (
	"errors"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
)

func GetReceivingChannel(cfg *config.Config) (<-chan *models.Knock, <-chan error, error) {
	if cfg.ReceiveRmq != nil {
		return initRmqReceiveChannel(cfg.ReceiveRmq)
	}
	return nil, nil, errors.New("no_receiver_configured")
}

func CloseReceiver(cfg *config.Config) {
	if cfg.ReceiveRmq != nil {
		closeRmqReceiver()
	}
}
