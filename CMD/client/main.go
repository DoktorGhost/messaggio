package main

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
	"strconv"
	"sync"
	"time"
)

// структура для сообщения
type MyMessage struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

func main() {
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Создание консьюмера Kafka
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Подписка на партицию "messages" в Kafka
	partitionConsumer, err := consumer.ConsumePartition("messages", 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	// Канал для завершения работы горутин
	var wg sync.WaitGroup

	// Обработка сообщений из Kafka
	for {
		select {
		case msg, ok := <-partitionConsumer.Messages():
			if !ok {
				log.Println("Channel closed, exiting")
				wg.Wait() // Ждем завершения всех горутин
				return
			}

			wg.Add(1)
			go func(msg *sarama.ConsumerMessage) {
				defer wg.Done()

				var message MyMessage
				err := json.Unmarshal(msg.Value, &message)
				if err != nil {
					log.Printf("Error unmarshaling JSON: %v\n", err)
					return
				}

				// Имитируем длительную обработку сообщения
				time.Sleep(5 * time.Second)

				responseText := strconv.Itoa(message.ID) + message.Content + " " + message.Status

				resp := &sarama.ProducerMessage{
					Topic: "responses",
					Key:   sarama.StringEncoder(strconv.Itoa(message.ID)),
					Value: sarama.StringEncoder(responseText),
				}

				_, _, err = producer.SendMessage(resp)
				if err != nil {
					log.Printf("Failed to send message to Kafka: %v", err)
				}
				log.Printf("Received message: %+v\n", message)
			}(msg)
		}
	}
}
