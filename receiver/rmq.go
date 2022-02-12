package receiver

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
)

var rmqConnection *amqp.Connection
var rmqChannel *amqp.Channel

func initRmqReceiveChannel(rmqConfig *config.RabbitmqConfig) (<-chan *models.Knock, <-chan error, error) {
	var err error
	var queue amqp.Queue

	rmqUrl := config.RmqUrl(rmqConfig)
	rmqConnection, err = amqp.Dial(rmqUrl)
	if err != nil {
		return nil, nil, err
	}

	rmqChannel, err = rmqConnection.Channel()
	if err != nil {
		return nil, nil, err
	}

	// all errors from rmw will be thrown in this channel
	errorChannel := rmqChannel.NotifyClose(make(chan *amqp.Error, 1))

	queue, err = rmqChannel.QueueDeclare(
		rmqConfig.Queue, // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, nil, err
	}

	consumeChannel, _ := rmqChannel.Consume(
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
			if err != nil {
				commonutils.ReportOnError(err, "receiver:: failed to unmarshal message")
			} else {
				deliveryChannel <- knock
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
	if rmqConnection != nil {
		_ = rmqConnection.Close()
	}
	if rmqChannel != nil {
		_ = rmqChannel.Close()
	}
}
