package produce

import (
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"time"
)

const (
	kafkaBrokers  = "124.222.86.11:9092"
	topic         = "go-kafka-demo"
	consumerGroup = "go-consumers"
)

// 生产者示例
func ProduceMessages() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true                   //如果启用，成功送达的消息将在“Successes”通道中返回（默认禁用）
	config.Producer.Partitioner = sarama.NewRandomPartitioner //随机分区测率

	producer, err := sarama.NewSyncProducer([]string{kafkaBrokers}, config)
	if err != nil {
		log.Fatalf("创建生产者失败：%v", err)
	}
	defer producer.Close()

	for i := 0; i < 10; i++ {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(fmt.Sprintf("消息 %d - %s", i, time.Now().Format(time.RFC3339))),
		}

		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("发送消息失败：%v", err)
		} else {
			log.Printf("发送消息成功！分区：%d,偏移量：%d", partition, offset)
		}
		time.Sleep(1 * time.Second)
	}
}
