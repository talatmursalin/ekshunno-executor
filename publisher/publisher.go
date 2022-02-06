package publisher

import (
	"fmt"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
	"os"
)

var file *os.File

func ConfigurePublisher(cfg *config.Config, publisherChannel <-chan *models.Result) {
	if cfg.PublishFile != nil {
		err := writeToFile(cfg.PublishFile, publisherChannel)
		commonutils.ReportOnError(err, "File publisher configure error")
	}
}

func writeToFile(fileConfig *config.PublishFileConfig, publisherChannel <-chan *models.Result) error {
	var err error
	file, err = os.OpenFile(fileConfig.Name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	go func() {
		result := <-publisherChannel
		data := []byte(fmt.Sprintf("Verdict: %s\tTime: %file\tMemory: %file\n", result.Verdict, result.Time, result.Memory))
		_, err = file.Write(data)
		if err != nil {
			fmt.Println("File write error:", err)
		}
	}()
	return nil
}

func ClosePublisher(_ *config.Config) {
	if file != nil {
		file.Close()
	}
}
