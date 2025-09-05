package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func main() {
	//1.建立连接
	conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
	failOnError(err, "failed to connect to rabbitMQ")
	defer conn.Close()

	//2.建立通道
	ch, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	defer ch.Close()

	//3.声明交换机
	err = ch.ExchangeDeclare(
		"logs_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "failed to declare an exchange")

	//4.声明队列
	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to declare a queue")

	if len(os.Args) < 2 {
		log.Printf("Usage: %s [info] [warning] [error]", os.Args[0])
		os.Exit(0)
	}

	for _, s := range os.Args[1:] {
		log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, "logs_direct", s)
		//5.绑定队列
		err = ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil)

		failOnError(err, "failed to bind a queue")
	}

	//6.消费队列
	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to register a consumer")

	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)

		}
	}()

	log.Printf("[*] waiting for messages, to exit press ctrl+c")
	<-forever
}
