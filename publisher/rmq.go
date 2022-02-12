package publisher

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"github.com/talatmursalin/ekshunno-executor/models"
)

var rmqConnection *amqp.Connection
var rmqChannel *amqp.Channel

func initRmqPublisher(rmqConfig *config.RabbitmqConfig) (chan<- *models.Result, <-chan error, error) {
	var err error
	var queue amqp.Queue // no need to declare queue when push to rmq

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

	publishChannel := make(chan *models.Result)
	errChan := make(chan error)

	go func() {
		for {
			result := <-publishChannel
			knock, err := json.Marshal(result)
			if err == nil {
				err = rmqChannel.Publish(
					"",
					queue.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        knock,
					})
				if err != nil {
					commonutils.OnError(err, "publisher:: failed to publish msg in queue")
				}
			} else {
				commonutils.ReportOnError(err, "publisher:: failed to unmarshal message")
			}
		}
	}()

	go func() {
		for {
			errChan <- <-errorChannel
		}
	}()

	return publishChannel, errChan, err
}

func closeRmqReceiver() {
	if rmqConnection != nil {
		_ = rmqConnection.Close()
	}
	if rmqChannel != nil {
		_ = rmqChannel.Close()
	}
}
