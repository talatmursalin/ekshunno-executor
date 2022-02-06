package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/customenums"
	"github.com/talatmursalin/ekshunno-executor/models"
	"github.com/talatmursalin/ekshunno-executor/publisher"
	"github.com/talatmursalin/ekshunno-executor/receiver"
	"github.com/talatmursalin/ekshunno-executor/xcore/executor"
	"log"
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
	log.Printf("result : [verdict: %s time:%0.3f sec memory: %0.2f mb]",
		result.Verdict, result.Time, result.Memory)
	verByte, _ := json.Marshal(result)
	verdictChannel <- verdictMessage{msg: verByte, room: knock.SubmissionRoom}
	done <- true
}

//
//func publishVerdict() {
//	for {
//		msg := <-verdictChannel
//		ch, err := rmqConnection.Channel()
//		commonutils.ExitOnError(err, "Failed to open a channel")
//		// defer ch.Close()
//
//		q, err := ch.QueueDeclare(
//			msg.room, // name
//			false,    // durable
//			false,    // delete when unused
//			false,    // exclusive
//			false,    // no-wait
//			nil,      // arguments
//		)
//		commonutils.ExitOnError(err, "Failed to declare a queue")
//
//		err = ch.Publish(
//			"",     // exchange
//			q.Name, // routing key
//			false,  // mandatory
//			false,  // immediate
//			amqp.Publishing{
//				ContentType: "text/plain",
//				Body:        []byte(msg.msg),
//			})
//		commonutils.ExitOnError(err, "Failed to publish a message")
//		log.Printf(" [x] Sent to %s", msg.room)
//	}
//}

//func initVeridctChannel() {
//	verdictChannel = make(chan verdictMessage)
//	go publishVerdict()
//}

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
	log.Printf(knockErr.Output)
	errByte, _ := json.Marshal(knockErr)
	return errByte
}

var AppConfig *config.Config

func main() {
	var err error
	AppConfig = config.LoadConfig("./config.yml")

	// setup receiver
	var msgChan <-chan *models.Knock
	var errorChannel <-chan error
	msgChan, errorChannel, err = receiver.GetReceivingChannel(AppConfig)
	defer receiver.CloseReceiver(AppConfig)
	commonutils.ExitOnError(err, "Failed to setup message queue")

	// setup publisher
	publishChannel := make(chan *models.Result)
	publisher.ConfigurePublisher(AppConfig, publishChannel)

	//
	initWorkerPool(AppConfig.Concurrency) // max three concurrent judge process
	//initVeridctChannel()
	forever := make(chan bool)
	go func() {
		for {
			select {
			case e := <-errorChannel:
				commonutils.ReportOnError(e, "Connection failed")
				msgChan, errorChannel, err = receiver.GetReceivingChannel(AppConfig)
				commonutils.ExitOnError(err, "Failed to setup message queue :: retry")
			case knock := <-msgChan:
				err := knock.Validate()
				if err != nil {
					//go publishVerdict()
					getSubmissionErrorResult(err)
				} else {
					freeWorker := <-workerPoolChannel
					go runSubmission(knock, publishChannel, freeWorker.from)
				}
			}
		}
	}()

	log.Printf(" [x] Waiting for messages. To exit press CTRL+C")
	defer receiver.CloseReceiver(AppConfig)
	defer publisher.ClosePublisher(AppConfig)
	<-forever
}
