package main

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func fib(n int) int {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else {
		return fib(n-1) + fib(n-2)
	}
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

	//3.声明通道
	q, err := ch.QueueDeclare(
		"rpc_queue",
		false,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "failed to declare an queue")

	//4.服务质量 控制服务器在收到交付确认之前，为消费者在网络中尝试保留的消息数量或字节数
	// 运行多个服务器进程，为了将负载均衡分布到多个服务器上，我们需要设置prefetch通道
	err = ch.Qos(
		1,
		0,
		false,
	)
	failOnError(err, "failed to set Qos")

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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for d := range msgs {
			n, err := strconv.Atoi(string(d.Body))
			failOnError(err, "failed to convert body to integer")

			log.Printf("[.]fib(%d)", n)
			response := fib(n)

			err = ch.PublishWithContext(
				ctx,
				"",
				d.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(strconv.Itoa(response)),
				})
			failOnError(err, "failed to publish a message")

			d.Ack(false)
		}
	}()

	log.Printf("[*] Awaiting RPC requests")
	<-forever
}
