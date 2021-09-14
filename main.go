package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
	"github.com/talatmursalin/ekshunno-executor/models"
	"github.com/talatmursalin/ekshunno-executor/xcore/executor"
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
	"gopkg.in/yaml.v2"
)

var (
	rmqConnection   *amqp.Connection
	receiverChannel *amqp.Channel
	errorChannel    chan *amqp.Error
	cfg             models.Config
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(err)
	}
}

func loadConfig() {
	f, err := os.Open("config.yml")
	failOnError(err, "Failed to read config.yaml")
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	failOnError(err, "Failed to decode config")
}

func runSubmission(knock models.Knock) utils.Result {
	lang, _ := utils.StringToLangId(knock.Submission.Lang)
	limit := utils.NewLimit(
		knock.Submission.Time,
		knock.Submission.Memory,
		0.25) // .1/4mb | 256kb
	sDec, _ := b64.StdEncoding.DecodeString(knock.Submission.Src)
	executor := executor.GetExecutor(lang, string(sDec), *limit)

	result := executor.Compile()
	if result.Verdict == utils.OK {
		result = executor.Execute(knock.Submission.Input)
	}
	return result
}

func pushMessageToQueue(msg []byte, queue string) {
	ch, err := rmqConnection.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent to %s", queue)
}

func setUpConsumer() (<-chan amqp.Delivery, error) {
	q, err := receiverChannel.QueueDeclare(
		cfg.Rabbitmq.Queue, // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

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

func setUpMessageQueue(url string) (<-chan amqp.Delivery, error) {

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

	return setUpConsumer()
}

func getSubmissionErrorResult(err error) []byte {
	knockErr := utils.Result{
		Verdict: utils.IS,
		Time:    0,
		Memory:  0,
		Output:  fmt.Sprintf("Invalid Submission: %s", err.Error()),
	}
	log.Printf(knockErr.Output)
	errByte, _ := json.Marshal(knockErr)
	return errByte
}

func main() {
	loadConfig()
	rmqUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.Rabbitmq.Username, cfg.Rabbitmq.Password,
		cfg.Rabbitmq.Host, cfg.Rabbitmq.Port)
	msgs, err := setUpMessageQueue(rmqUrl)
	failOnError(err, "Failed to setup message queue")
	defer rmqConnection.Close()
	defer receiverChannel.Close()
	forever := make(chan bool)
	go func() {
		for {
			select {
			case e := <-errorChannel:
				log.Printf("Connection failed: %s", e.Error())
				for {
					time.Sleep(5 * time.Second)
					msgs, err = setUpMessageQueue(rmqUrl)
					if err != nil {
						log.Printf("Failed to reconnect: %s", err)
						continue
					}
					log.Printf("[x] Reconnected to queue")
					break
				}
			case msg := <-msgs:
				var msgByte []byte
				knock := models.Knock{}
				err := json.Unmarshal(msg.Body, &knock)
				if err != nil {
					msgByte = getSubmissionErrorResult(err)
				} else {
					err := knock.Validate()
					if err != nil {
						msgByte = getSubmissionErrorResult(err)
					} else {
						result := runSubmission(knock)
						log.Printf("result : [verdict: %s time:%0.3f sec memory: %0.2f mb]",
							result.Verdict, result.Time, result.Memory)
						msgByte, _ = json.Marshal(result)

					}

				}
				pushMessageToQueue(msgByte, knock.SubmissionRoom)
			}
		}
	}()

	log.Printf(" [x] Waiting for messages. To exit press CTRL+C")
	<-forever
}
