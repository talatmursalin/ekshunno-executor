package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
	"github.com/talatmursalin/ekshunno-executor/models"
	"github.com/talatmursalin/ekshunno-executor/xcore/executor"
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
	"gopkg.in/yaml.v2"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(err)
	}
}

func runSubmission(knock models.Knock) utils.Result {
	lang := utils.StringToLangId(knock.Submission.Lang)
	limit := utils.NewLimit(knock.Submission.Time, knock.Submission.Memory, 10)
	sDec, _ := b64.StdEncoding.DecodeString(knock.Submission.Src)
	executor := executor.GetExecutor(lang, string(sDec), *limit)

	result := executor.Compile()
	if result.Verdict == utils.OK {
		result = executor.Execute(knock.Submission.Input)
	}
	return result
}

func setUpConsumer(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	q, err := ch.QueueDeclare(
		"submission_queue", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	return ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func pushMessageToQueue(conn *amqp.Connection, msg []byte, queue string) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
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

func loadConfig() models.Config {
	f, err := os.Open("config.yml")
	failOnError(err, "Failed to read config.yaml")
	defer f.Close()
	var cfg models.Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	failOnError(err, "Failed to decode config")
	return cfg
}

func main() {
	cfg := loadConfig()
	rmqUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.Rabbitmq.Username, cfg.Rabbitmq.Password,
		cfg.Rabbitmq.Host, cfg.Rabbitmq.Port)
	conn, err := amqp.Dial(rmqUrl)
	failOnError(err, "Could not connect to RMQ")
	defer conn.Close()
	recChan, err := conn.Channel()
	failOnError(err, "Could not connect to RMQ")
	defer recChan.Close()
	msgs, err := setUpConsumer(conn, recChan)
	failOnError(err, "Failed to set up consumer")
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			knock := models.Knock{}
			json.Unmarshal(d.Body, &knock)
			result := runSubmission(knock)
			log.Printf("result : [verdict: %s time:%0.3f sec memory: %0.2f mb]",
				result.Verdict, result.Time, result.Memory)
			resultByte, _ := json.Marshal(result)
			pushMessageToQueue(conn, resultByte, knock.SubmissionRoom)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
