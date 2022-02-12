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
	"time"
)

type workerPoolMsg struct {
	msg  bool
	from chan bool
}

var (
	workerPoolChannel chan workerPoolMsg
)

var (
	publishChannel chan<- *models.Result
	pubErrNotifier <-chan error
	msgChan        <-chan *models.Knock
	recErrNotifier <-chan error
)

const MAX_RETRY = 10

func executeSubmission(knock *models.Knock) models.Result {
	lang, _ := customenums.StringToLangId(knock.Submission.Lang)
	limit := models.NewLimit(
		knock.Submission.Time,
		knock.Submission.Memory,
		0.25) // .1/4mb | 256kb
	sDec, _ := b64.StdEncoding.DecodeString(knock.Submission.Src)
	theExecutor := executor.GetExecutor(lang, string(sDec), *limit)

	result := theExecutor.Compile()
	if result.Verdict == customenums.OK {
		result = theExecutor.Execute(knock.Submission.Input)
	}
	return result
}

func runSubmission(knock *models.Knock, publisherChannel chan<- *models.Result, done chan<- bool) {
	result := executeSubmission(knock)
	publisherChannel <- &result
	log.Debug().Msgf("result : [verdict: %s time:%0.3f sec memory: %0.2f mb]",
		result.Verdict, result.Time, result.Memory)
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

func connectReceiver() {
	var err error
	msgChan, recErrNotifier, err = receiver.GetReceivingChannel(AppConfig)
	retry_count := 1
	for err != nil && retry_count < MAX_RETRY {
		commonutils.OnError(err, fmt.Sprintf("main:: retry attempt: %d - failed to setup receiver", retry_count))
		time.Sleep(time.Second)
		msgChan, recErrNotifier, err = receiver.GetReceivingChannel(AppConfig)
		retry_count += 1
	}
	if err != nil && retry_count >= MAX_RETRY {
		panic("failed to connect with receiver")
	}
}

func connectPublisher() {
	var err error
	publishChannel, pubErrNotifier, err = publisher.ConfigurePublisher(AppConfig)
	retry_count := 1
	for err != nil && retry_count < MAX_RETRY {
		commonutils.OnError(err, fmt.Sprintf("main:: retry attempt: %d - failed to setup publisher", retry_count))
		time.Sleep(time.Second)
		publishChannel, pubErrNotifier, err = publisher.ConfigurePublisher(AppConfig)
		retry_count += 1
	}
	if err != nil && retry_count >= MAX_RETRY {
		panic("failed to connect with publisher")
	}
}

func main() {
	// set initial logger to console. this is necessary to report
	// config parsing error. we will config our global logger gain when
	// we have log config from yml
	logger.InitLogger()

	var err error
	AppConfig, err = config.LoadConfig("./config.yml")
	if err != nil {
		panic(err)
	}
	// config logger
	err = logger.ConfigureLogger(AppConfig)
	if err != nil {
		commonutils.OnError(err, "logger can not be configured")
	}

	// setup connections
	connectReceiver()
	connectPublisher()

	initWorkerPool(AppConfig.Concurrency) // max concurrent judge process
	forever := make(chan bool)
	go func() {
		for {
			select {
			case e := <-recErrNotifier:
				commonutils.OnError(e, "main:: connection interrupted for receiver")
				connectReceiver()
			case e := <-pubErrNotifier:
				commonutils.OnError(e, "main:: connection interrupted for publisher")
				connectPublisher()
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
