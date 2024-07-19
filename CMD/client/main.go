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

	var producer sarama.SyncProducer
	var consumer sarama.Consumer
	var partitionConsumer sarama.PartitionConsumer
	var err error
	initialDelay := 5 * time.Second

	for i := 1; i < 6; i++ {
		producer, err = sarama.NewSyncProducer([]string{"kafka:9092"}, nil)
		if err != nil {
			if i < 5 {
				log.Printf("Создание producer. Попытка №%d не удалось: %v", i, err)
				log.Printf("Повторная попытка создания producer")
				time.Sleep(initialDelay * time.Duration(i))
				continue
			} else {
				log.Fatal("Ошибка создания producer: %v", err)
			}
		} else {
			log.Printf("producer успешно создан")
			break
		}
	}
	defer producer.Close()

	// Создание консьюмера Kafka
	for i := 1; i < 6; i++ {
		consumer, err = sarama.NewConsumer([]string{"kafka:9092"}, nil)
		if err != nil {
			if i < 5 {
				log.Printf("Создание consumer. Попытка №%d не удалось: %v", i, err)
				log.Printf("Повторная попытка создания consumer")
				time.Sleep(initialDelay * time.Duration(i))
				continue
			} else {
				log.Fatal("Ошибка создания consumer: %v", err)
			}
		} else {
			log.Printf("consumer успешно создан")
			break
		}
	}
	defer consumer.Close()

	// Подписка на партицию "messages" в Kafka
	for i := 1; i < 6; i++ {
		partitionConsumer, err = consumer.ConsumePartition("messages", 0, sarama.OffsetOldest)
		if err != nil {
			if i < 5 {
				log.Printf("Подписка partition. Попытка №%d не удалось: %v", i, err)
				log.Printf("Повторная попытка создания partition")
				time.Sleep(initialDelay * time.Duration(i))
				continue
			} else {
				log.Fatal("Ошибка partition: %v", err)
			}
		} else {
			log.Printf("Подписка на partition успешно выполнена")
			break
		}
	}
	defer partitionConsumer.Close()

	// Канал для завершения работы горутин
	var wg sync.WaitGroup

	// Обработка сообщений из Kafka
	for {
		select {
		case msg, ok := <-partitionConsumer.Messages():
			if !ok {
				log.Println("Канал закрыт")
				wg.Wait() // Ждем завершения всех горутин
				return
			}

			wg.Add(1)
			go func(msg *sarama.ConsumerMessage) {
				defer wg.Done()

				var message MyMessage
				err := json.Unmarshal(msg.Value, &message)
				if err != nil {
					log.Printf("Ошибка JSON: %v\n", err)
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
					log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
				}
				log.Printf("Сообщение обработано: %+v\n", message)
			}(msg)
		}
	}
}
