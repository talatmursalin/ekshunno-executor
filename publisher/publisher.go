package publisher

import (
	"errors"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
)

func ConfigurePublisher(cfg *config.Config) chan *models.Result {
	if cfg.PublishRmq != nil {
		err := errors.New("publisher_not_configured")
		commonutils.ReportOnError(err, "File publisher configure error")
	}
	return make(chan *models.Result)
}

func ClosePublisher(_ *config.Config) {
	//
}
