package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/customenums"
	"github.com/talatmursalin/ekshunno-executor/logger"
	"github.com/talatmursalin/ekshunno-executor/models"
	"github.com/talatmursalin/ekshunno-executor/publisher"
	"github.com/talatmursalin/ekshunno-executor/receiver"
	"github.com/talatmursalin/ekshunno-executor/xcore/executor"
)

type workerPoolMsg struct {
	msg  bool
	from chan bool
}

type verdictMessage struct {
	msg  []byte
	room string
}

var (
	// workerDoneChannels []chan bool
	verdictChannel    chan verdictMessage
	workerPoolChannel chan workerPoolMsg
)

func executeSubmission(knock *models.Knock) models.Result {
	lang, _ := customenums.StringToLangId(knock.Submission.Lang)
	limit := models.NewLimit(
		knock.Submission.Time,
		knock.Submission.Memory,
		0.25) // .1/4mb | 256kb
	sDec, _ := b64.StdEncoding.DecodeString(knock.Submission.Src)
	executor := executor.GetExecutor(lang, string(sDec), *limit)

	result := executor.Compile()
	if result.Verdict == customenums.OK {
		result = executor.Execute(knock.Submission.Input)
	}
	return result
}

func runSubmission(knock *models.Knock, publisherChannel chan<- *models.Result, done chan<- bool) {
	result := executeSubmission(knock)
	publisherChannel <- &result
	log.Debug().Msgf("result : [verdict: %s time:%0.3f sec memory: %0.2f mb]",
		result.Verdict, result.Time, result.Memory)
	verByte, _ := json.Marshal(result)
	verdictChannel <- verdictMessage{msg: verByte, room: knock.SubmissionRoom}
	done <- true
}

func initWorkerPool(n int) {
	workerPoolChannel = make(chan workerPoolMsg, n)
	// workerDoneChannels = make([]chan bool, n)
	for i := 0; i < n; i++ {
		ch := make(chan bool, 1)
		ch <- true
		go func(ch chan bool) {
			for {
				msg := <-ch
				workerPoolChannel <- workerPoolMsg{msg: msg, from: ch}
			}
		}(ch)
		// log.Println("worker:", i)
		// workerDoneChannels[i] <- true
	}
}

//func retryConnection(url string) (<-chan amqp.Delivery, error) {
//	cnt := 0
//	for {
//		cnt += 1
//		time.Sleep(5 * time.Second)
//		msgs, err := initReceiveChannel(url)
//		if err != nil {
//			if cnt > 720 {
//				commonutils.ExitOnError(err, "Retry limit exceeded")
//			}
//			log.Printf("Failed to reconnect: %s", err)
//			continue
//		}
//		log.Printf("[x] Reconnected to queue")
//		return msgs, err
//	}
//}

func getSubmissionErrorResult(err error) []byte {
	knockErr := models.Result{
		Verdict: customenums.IS,
		Time:    0,
		Memory:  0,
		Output:  fmt.Sprintf("Invalid Submission: %s", err.Error()),
	}
	log.Debug().Msgf(knockErr.Output)
	errByte, _ := json.Marshal(knockErr)
	return errByte
}

var AppConfig *config.Config

func main() {
	// set initial logger to console. this is necessary to report
	// config parsing error. we will config our global logger gain when
	// we have log config from yml
	logger.InitLogger()

	var err error
	AppConfig = config.LoadConfig("./config.yml")
	// config logger
	_ = logger.ConfigureLogger(AppConfig)

	// setup receiver
	var msgChan <-chan *models.Knock
	var errorChannel <-chan error
	msgChan, errorChannel, err = receiver.GetReceivingChannel(AppConfig)
	defer receiver.CloseReceiver(AppConfig)
	if err != nil {
		commonutils.ReportOnError(err, "main:: failed to setup receiver")
		panic(err)
	}

	// setup publisher
	var publishChannel chan<- *models.Result
	publishChannel, _, err = publisher.ConfigurePublisher(AppConfig)
	if err != nil {
		commonutils.ReportOnError(err, "main:: failed to setup publisher")
		//panic(err) disabling exit on publisher error
	}
	//
	initWorkerPool(AppConfig.Concurrency) // max three concurrent judge process
	//initVeridctChannel()
	forever := make(chan bool)
	go func() {
		for {
			select {
			case e := <-errorChannel:
				commonutils.ReportOnError(e, "main:: connection failed")
				msgChan, errorChannel, err = receiver.GetReceivingChannel(AppConfig)
				if err != nil {
					commonutils.ReportOnError(err, "main:: retry failed to setup receiver")
					panic(err)
				}
			case knock := <-msgChan:
				err := knock.Validate()
				if err != nil {
					getSubmissionErrorResult(err)
				} else {
					freeWorker := <-workerPoolChannel
					go runSubmission(knock, publishChannel, freeWorker.from)
				}
			}
		}
	}()
	log.Info().Msgf("[x] waiting for messages. To exit press CTRL+C")
	defer receiver.CloseReceiver(AppConfig)
	defer publisher.ClosePublisher(AppConfig)
	<-forever
}
