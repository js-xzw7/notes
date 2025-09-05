package main

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strings"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}
func main() {
	//1.连接rabbitmq服务器
	conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
	failOnError(err, "Failed to connect to rabbitmq")
	defer conn.Close()

	//2.创建通道，大部分用于完成操作的api都驻留在通道中
	ch, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	defer ch.Close()

	//3.声明要发送的队列,声明队列是幂等的，只有当队列不存在时才会创建
	q, err := ch.QueueDeclare(
		"hello",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := bodyFrom(os.Args)
	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})

	failOnError(err, "failed to publish a message")
	log.Printf("[x] sent %s \n", body)
}
