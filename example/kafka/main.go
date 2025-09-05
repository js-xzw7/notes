package main

import (
	"github.com/IBM/sarama"
	"kafka/consume"
	"kafka/produce"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	kafkaBrokers  = "124.222.86.11:9092"
	topic         = "go-kafka-demo"
	consumerGroup = "go-consumers"
)

func main() {
	//创建主题
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	admin, err := sarama.NewClusterAdmin([]string{kafkaBrokers}, config)
	if err == nil {
		err = admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     3,
			ReplicationFactor: 1,
		}, false)

		if err != nil {
			log.Panicf("创建主题失败：%v （可能已经存在）", err)
		}

		admin.Close()
	}

	//启动消费者组（在后台运行）
	go consume.ConsumeWithGroup()
	time.Sleep(2 * time.Second) //等待消费者组准备

	//运行生产者
	go produce.ProduceMessages()
	time.Sleep(12 * time.Second) //等待消息发送完成

	//运行事务生产者
	go consume.TransactionalProducer()
	time.Sleep(3 * time.Second)

	//启动独立消费者
	go consume.ConsumeMessages()

	//等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("程序退出")
}
