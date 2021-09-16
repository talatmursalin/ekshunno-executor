package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/customenums"
	"github.com/talatmursalin/ekshunno-executor/models"
	"github.com/talatmursalin/ekshunno-executor/xcore/executor"
	"gopkg.in/yaml.v2"
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
	rmqConnection     *amqp.Connection
	receiverChannel   *amqp.Channel
	errorChannel      chan *amqp.Error
	cfg               models.Config
	workerPoolChannel chan workerPoolMsg
	// workerDoneChannels []chan bool
	verdictChannel chan verdictMessage
)

func loadConfig() {
	f, err := os.Open("config.yml")
	commonutils.ExitOnError(err, "Failed to read config.yaml")
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	commonutils.ExitOnError(err, "Failed to decode config")
}

func executeSubmission(knock models.Knock) models.Result {
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

func runSubmission(knock models.Knock, done chan bool) {
	result := executeSubmission(knock)
	log.Printf("result : [verdict: %s time:%0.3f sec memory: %0.2f mb]",
		result.Verdict, result.Time, result.Memory)
	verByte, _ := json.Marshal(result)
	verdictChannel <- verdictMessage{msg: verByte, room: knock.SubmissionRoom}
	done <- true
}

func publishVerdict() {
	for {
		msg := <-verdictChannel
		ch, err := rmqConnection.Channel()
		commonutils.ExitOnError(err, "Failed to open a channel")
		// defer ch.Close()

		q, err := ch.QueueDeclare(
			msg.room, // name
			false,    // durable
			false,    // delete when unused
			false,    // exclusive
			false,    // no-wait
			nil,      // arguments
		)
		commonutils.ExitOnError(err, "Failed to declare a queue")

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(msg.msg),
			})
		commonutils.ExitOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent to %s", msg.room)
	}
}

func initConsumer() (<-chan amqp.Delivery, error) {
	q, err := receiverChannel.QueueDeclare(
		cfg.Rabbitmq.Queue, // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	commonutils.ExitOnError(err, "Failed to declare a queue")

	return receiverChannel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func initReceiveChannel(url string) (<-chan amqp.Delivery, error) {

	var err error
	rmqConnection, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	receiverChannel, err = rmqConnection.Channel()
	if err != nil {
		return nil, err
	}

	closeChan := make(chan *amqp.Error, 1)
	errorChannel = receiverChannel.NotifyClose(closeChan)

	return initConsumer()
}

func initVeridctChannel() {
	verdictChannel = make(chan verdictMessage)
	go publishVerdict()
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

func retryConnection(url string) (<-chan amqp.Delivery, error) {
	cnt := 0
	for {
		cnt += 1
		time.Sleep(5 * time.Second)
		msgs, err := initReceiveChannel(url)
		if err != nil {
			if cnt > 720 {
				commonutils.ExitOnError(err, "Retry limit exceeded")
			}
			log.Printf("Failed to reconnect: %s", err)
			continue
		}
		log.Printf("[x] Reconnected to queue")
		return msgs, err
	}
}

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

func getRmqURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.Rabbitmq.Username, cfg.Rabbitmq.Password,
		cfg.Rabbitmq.Host, cfg.Rabbitmq.Port)
}

func main() {
	loadConfig()
	rmqUrl := getRmqURL()
	msgs, err := initReceiveChannel(rmqUrl)
	commonutils.ExitOnError(err, "Failed to setup message queue")
	defer rmqConnection.Close()
	defer receiverChannel.Close()
	initWorkerPool(cfg.General.Concurrency) // max three concurrent judge process
	initVeridctChannel()
	forever := make(chan bool)
	go func() {
		for {
			select {
			case e := <-errorChannel:
				commonutils.ReportOnError(e, "Connection failed")
				msgs, err = retryConnection(rmqUrl)
			case msg := <-msgs:
				knock := models.Knock{}
				err := json.Unmarshal(msg.Body, &knock)
				if err != nil {
					subErr := getSubmissionErrorResult(err)
					commonutils.ReportOnError(err, string(subErr))
				} else {
					err := knock.Validate()
					if err != nil {
						go publishVerdict()
						getSubmissionErrorResult(err)
					} else {
						freeWorker := <-workerPoolChannel
						go runSubmission(knock, freeWorker.from)
					}

				}
			}
		}
	}()

	log.Printf(" [x] Waiting for messages. To exit press CTRL+C")
	<-forever
}
