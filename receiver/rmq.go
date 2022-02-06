package receiver

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
)

var rmqConnection *amqp.Connection
var receiverChannel *amqp.Channel

func initRmqReceiveChannel(rmqConfig *config.RabbitmqConfig) (<-chan *models.Knock, <-chan error, error) {
	var err error
	var queue amqp.Queue

	rmqUrl := config.RmqUrl(rmqConfig)
	rmqConnection, err = amqp.Dial(rmqUrl)
	if err != nil {
		return nil, nil, err
	}

	receiverChannel, err = rmqConnection.Channel()
	if err != nil {
		return nil, nil, err
	}

	// all errors from rmw will be thrown in this channel
	errorChannel := receiverChannel.NotifyClose(make(chan *amqp.Error, 1))

	queue, err = receiverChannel.QueueDeclare(
		rmqConfig.Queue, // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	commonutils.ExitOnError(err, "Failed to declare a queue")

	consumeChannel, _ := receiverChannel.Consume(
		queue.Name,  // queue
		"go_client", // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)

	deliveryChannel := make(chan *models.Knock)
	errChan := make(chan error)

	go func() {
		for {
			msg := <-consumeChannel
			knock := &models.Knock{}
			err := json.Unmarshal(msg.Body, &knock)
			if err == nil {
				deliveryChannel <- knock
			} else {
				fmt.Println("Message can not be unmarshal-ed", err)
			}
		}
	}()

	go func() {
		for {
			errChan <- <-errorChannel
		}
	}()

	return deliveryChannel, errChan, err
}

func closeRmqReceiver() {
	_ = rmqConnection.Close()
	_ = receiverChannel.Close()
}
