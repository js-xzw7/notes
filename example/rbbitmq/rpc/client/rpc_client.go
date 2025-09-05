package main

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func fibonacciRPC(n int) (res int, err error) {
	//1.建立连接
	conn, err := amqp.Dial("amqp://guest:guest@124.222.86.11:5672/")
	failOnError(err, "failed to connect to rabbitMQ")
	defer conn.Close()

	//2.声明通道
	ch, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	defer ch.Close()

	//3.声明队列
	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	failOnError(err, "failed to declare a queue")

	//4.消费
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to register a consumer")

	corrId := randomString(32)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(strconv.Itoa(n)),
		})
	failOnError(err, "failed to publish a message")

	for d := range msgs {
		if corrId == d.CorrelationId {
			res, err = strconv.Atoi(string(d.Body))
			failOnError(err, "failed to convert body to integer")
			break
		}
	}
	return
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	n := bodyFrom(os.Args)

	log.Printf(" [x] Requesting fib(%d)", n)
	res, err := fibonacciRPC(n)
	failOnError(err, "failed to handle rpc request")
	log.Printf(" [.] Got %d", res)
}

func bodyFrom(args []string) int {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "30"
	} else {
		s = strings.Join(args[1:], " ")
	}

	n, err := strconv.Atoi(s)
	failOnError(err, "failed to convert arg to integer")
	return n
}
