package consume

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	kafkaBrokers  = "124.222.86.11:9092"
	topic         = "go-kafka-demo"
	consumerGroup = "go-consumers"
)

// 消费者示例
func ConsumeMessages() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{kafkaBrokers}, config)
	if err != nil {
		log.Fatalf("创建消费者失败")
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalf("创建分区消费者失败：%v", err)
	}
	defer partitionConsumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

consumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("收到消息：分区：%d, 偏移量：%d,内容：%s", msg.Partition, msg.Offset, string(msg.Value))
		case err := <-partitionConsumer.Errors():
			log.Printf("消费错误：%v", err)
		case <-signals:
			break consumerLoop
		}
	}
}

// 消费者组处理器
type consumerGroupHandler struct{}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("消费者组消费消息：主题：%s, 分区：%d,偏移量：%d,内容：%s", msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		//标记消息已处理
		sess.MarkMessage(msg, "")
	}
	return nil
}

// 消费者组示例
func ConsumeWithGroup() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup([]string{kafkaBrokers}, consumerGroup, config)
	if err != nil {
		log.Fatalf("创建消费者组失败：%v", err)
	}
	defer group.Close()

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := group.Consume(ctx, []string{topic}, &consumerGroupHandler{}); err != nil {
				log.Printf("消费错误：%v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGTERM)
	<-sigterm
	cancel()
	wg.Wait()
}

// 事务生产者示例
func TransactionalProducer() {
	config := sarama.NewConfig()
	config.Producer.Idempotent = true
	config.Producer.Transaction.ID = "go-transactional-producer"
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer([]string{kafkaBrokers}, config)
	if err != nil {
		log.Fatalf("创建事务生产者失败：%v", err)
	}
	defer producer.Close()

	//初始化事务
	err = producer.BeginTxn()
	if err != nil {
		log.Fatalf("开始事务失败：%v", err)
	}

	//发送事务消息
	for i := 0; i < 5; i++ {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(fmt.Sprintf("事务消息%d", i)),
		}

		_, _, err = producer.SendMessage(msg)
		if err != nil {
			log.Fatalf("发送事务消息失败：%v", err)

			if err2 := producer.AbortTxn(); err2 != nil {
				log.Printf("终止事务失败：%v", err)
			}
			return
		}
	}

	//提交事务
	if err := producer.CommitTxn(); err != nil {
		log.Printf("提交事务失败：%v", err)
	} else {
		log.Println("提交事务成功")
	}
}
