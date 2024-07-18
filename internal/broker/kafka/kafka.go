package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"messaggio/internal/config"
	"messaggio/internal/models"
)

type KafkaBroker struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func NewKafkaBroker(conf *config.Config) (*KafkaBroker, error) {

	// Создание продюсера Kafka
	producer, err := sarama.NewSyncProducer([]string{"kafka:9092"}, nil)
	if err != nil {
		log.Printf("Ошибка создания producer: %v", err)
		return nil, err
	}

	// Создание консьюмера Kafka
	consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, nil)
	if err != nil {
		producer.Close()
		log.Printf("Ошибка создания consumer: %v", err)
		return nil, err
	}

	broker := &KafkaBroker{
		producer: producer,
		consumer: consumer,
	}

	return broker, nil

}

func (kb *KafkaBroker) ConsumeResponses(ctx context.Context, messageIDChannel chan<- string) error {
	// Подписка на партицию "responses" в Kafka
	partitionConsumer, err := kb.consumer.ConsumePartition("responses", 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("Ошибка партиции: %v", err)
	}

	go func() {
		for {
			select {
			case msg, ok := <-partitionConsumer.Messages():
				if !ok {
					log.Println("Канал закрыт")
					return
				}
				responseID := string(msg.Key)
				messageIDChannel <- responseID
			}
		}
	}()

	return nil
}

func (kb *KafkaBroker) SendMessage(message models.Message) error {
	value, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "messages",
		Value: sarama.StringEncoder(value),
	}

	_, _, err = kb.producer.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

func (kb *KafkaBroker) Close() {
	if kb.producer != nil {
		kb.producer.Close()
	}
	if kb.consumer != nil {
		kb.consumer.Close()
	}
}
