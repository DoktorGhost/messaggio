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

/*
func NewKafkaBroker(conf *config.Config) (*KafkaBroker, error) {
	// Инициализация Kafka producer
	configKafka := sarama.NewConfig()
	configKafka.Producer.RequiredAcks = sarama.WaitForAll
	configKafka.Producer.Retry.Max = 5
	configKafka.Producer.Return.Successes = true

	// Создание продюсера Kafka
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, configKafka)
	if err != nil {
		//log.Fatalf("Failed to create producer: %v", err)
		return nil, err
	}

	// Создание консьюмера Kafka
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, configKafka)
	if err != nil {
		producer.Close()
		//log.Fatalf("Failed to create consumer: %v", err)
		return nil, err
	}

	// Проверяем и создаем необходимые топики
	admin, err := sarama.NewClusterAdmin([]string{"localhost:9092"}, configKafka)
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}
	defer admin.Close()

	// Создаем топики, если они еще не существуют
	err = createTopicIfNotExists(admin, "messages")
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}

	err = createTopicIfNotExists(admin, "responses")
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}

	broker := &KafkaBroker{
		producer: producer,
		consumer: consumer,
	}

	return broker, nil

}

func createTopicIfNotExists(admin sarama.ClusterAdmin, topic string) error {
	topics, err := admin.ListTopics()
	if err != nil {
		return err
	}

	if _, exists := topics[topic]; !exists {
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}

		err := admin.CreateTopic(topic, topicDetail, false)
		if err != nil {
			return err
		}

		//log.Printf("Created topic '%s'", topic)
	}

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

func (kb *KafkaBroker) ConsumeResponses(ctx context.Context, messageIDChannel chan<- int) error {
	// Подписка на партицию "responses" в Kafka
	partitionConsumer, err := kb.consumer.ConsumePartition("responses", 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				var response string
				log.Printf("тут ошибка1: %s\n", msg)
				err := json.Unmarshal(msg.Value, &response)
				log.Printf("тут ошибка2: %s\n", response)
				if err != nil {
					log.Printf("Error unmarshaling JSON: %v\n", err)
					continue
				}

				log.Printf("Received response: %s\n", response)

				messageIDStr := string(msg.Value)
				messageID, err := strconv.Atoi(messageIDStr)
				if err != nil {
					log.Printf("Error converting message ID to int: %v\n", err)
				}

				// Отправляем ID в канал для обновления в БД
				messageIDChannel <- messageID

			case <-ctx.Done():
				log.Println("Context canceled, stopping response consumption")
				return
			}
		}
	}()

	return nil
}
*/

func NewKafkaBroker(conf *config.Config) (*KafkaBroker, error) {

	// Создание продюсера Kafka
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
		return nil, err
	}

	// Создание консьюмера Kafka
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		producer.Close()
		log.Fatalf("Failed to create consumer: %v", err)
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
		return fmt.Errorf("failed to consume partition: %v", err)
	}

	go func() {
		for {
			select {
			case msg, ok := <-partitionConsumer.Messages():
				if !ok {
					log.Println("Channel closed, exiting goroutine")
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
