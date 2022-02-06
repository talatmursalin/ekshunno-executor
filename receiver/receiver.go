package receiver

import (
	"errors"
	"fmt"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
)

func GetReceivingChannel(cfg *config.Config) (<-chan *models.Knock, <-chan error, error) {
	if cfg.ReceiveRmq != nil {
		return initRmqReceiveChannel(cfg.ReceiveRmq)
	}
	return nil, nil, errors.New("No receiver configured")
}

func CloseReceiver(cfg *config.Config) {
	fmt.Println("Closing Receiver")
	if cfg.ReceiveRmq != nil {
		closeRmqReceiver()
	}
}
