package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

var (
	uri           string
	exchange      string
	exchangeType  string
	queue         string
	bindingKey    string
	consumerTag   string
	lifetime      time.Duration
	verbose       bool
	autoAck       bool
	ErrLog        = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmsgprefix)
	Log           = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmsgprefix)
	deliveryCount int
)

func init() {
	viper.AutomaticEnv()

	viper.SetDefault("URI", "amqp://demo:demo@localhost:5672/")
	viper.SetDefault("EXCHANGE", "scalable")
	viper.SetDefault("QUEUE", "test-queue")
	viper.SetDefault("BINDING_KEY", "")
	viper.SetDefault("CONSUMER_TAG", "test-consumer")
	// viper.SetDefault("LIFETIME", 30*time.Second)
	viper.SetDefault("VERBOSE", true)
	viper.SetDefault("AUTO_ACK", false)

	uri = viper.GetString("URI")
	exchange = viper.GetString("EXCHANGE")
	queue = viper.GetString("QUEUE")
	bindingKey = viper.GetString("BINDING_KEY")
	consumerTag = viper.GetString("CONSUMER_TAG")
	// lifetime = viper.GetDuration("LIFETIME")
	verbose = viper.GetBool("VERBOSE")
	autoAck = viper.GetBool("AUTO_ACK")
}

func main() {
	c, err := NewConsumer(uri, exchange, queue, bindingKey, consumerTag)
	if err != nil {
		ErrLog.Fatalf("%s", err)
	}

	SetupCloseHandler(c)

	Log.Printf("running consumer")
	c.Consume(queue)

	// if lifetime > 0 {
	// 	Log.Printf("running for %s", lifetime)
	// 	time.Sleep(lifetime)
	// } else {
	// 	Log.Printf("running until Consumer is done")
	// 	<-c.done
	// }

	// Log.Printf("shutting down")

	// if err := c.Shutdown(); err != nil {
	// 	ErrLog.Fatalf("error during shutdown: %s", err)
	// }
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func SetupCloseHandler(consumer *Consumer) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// <-c
	// Log.Printf("Ctrl+C pressed in Terminal")
	// if err := consumer.Shutdown(); err != nil {
	// 	ErrLog.Fatalf("error during shutdown: %s", err)
	// }
	// os.Exit(0)
	go func() {
		<-c
		Log.Printf("Ctrl+C pressed in Terminal")
		if err := consumer.Shutdown(); err != nil {
			ErrLog.Fatalf("error during shutdown: %s", err)
		}
		os.Exit(0)
	}()
}

func NewConsumer(amqpURI, exchange, queueName, key, ctag string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error

	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	Log.Printf("dialing %q", amqpURI)
	c.conn, err = amqp.DialConfig(amqpURI, config)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		Log.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	Log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	Log.Printf("declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	Log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	return c, nil
}

func (c *Consumer) Consume(queue string) error {
	Log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	for {
		deliveries, err := c.channel.Consume(
			queue,   // name
			c.tag,   // consumerTag,
			autoAck, // autoAck
			false,   // exclusive
			false,   // noLocal
			false,   // noWait
			nil,     // arguments
		)
		if err != nil {
			return fmt.Errorf("Queue Consume: %s", err)
		}

		handle(deliveries, c.done)
	}
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer Log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	// cleanup := func() {
	// 	Log.Printf("handle: deliveries channel closed")
	// 	done <- nil
	// }

	// defer cleanup()

	for d := range deliveries {
		deliveryCount++
		if verbose {
			Log.Printf(
				"got %dB delivery: [%v] %q",
				len(d.Body),
				d.DeliveryTag,
				d.Body,
			)
		} else {
			if deliveryCount%65536 == 0 {
				Log.Printf("delivery count %d", deliveryCount)
			}
		}
		time.Sleep(10 * time.Millisecond)
		if !autoAck {
			d.Ack(false)
		}
	}
}
