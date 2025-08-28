package myKafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"l0/internal/cache"
	"l0/internal/models"
	"l0/internal/service"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

var kafkaReader = kafka.NewReader(kafka.ReaderConfig{
	Brokers: []string{os.Getenv("KAFKA_URL")},
	Topic:   "orders",
})
var kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
	Brokers: []string{os.Getenv("KAFKA_URL")},
	Topic:   "orders",
})

func GetOrder(db *sql.DB, cache *cache.TTLMap) {
	var order models.Order
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader := kafkaReader
	defer reader.Close()
	for {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Println("Ошибка при прочтении сообщения: ", err)

		} else {
			err := json.Unmarshal(message.Value, &order)
			if err != nil {
				log.Println("Не удалось распарсить JSON: ", err)
				continue
			}
		}
		err = service.InsertToDB(db, order)
		if err != nil {
			log.Printf("Ошибка вставки в БД: %v", err)
			continue
		}
		cache.Set(order.OrderUID, order)
		kafkaReader.CommitMessages(ctx, message)
		log.Printf("Получен заказ: %v", order.OrderUID)
	}
}
func SendOrder(orders int, ch chan models.Order) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	writer := kafkaWriter
	defer writer.Close()
	log.Println("Создание сообщений для кафки...")
	for i := 0; i < orders; i++ {
		order := <-ch
		jsonOrder, err := json.Marshal(order)
		if err != nil {
			log.Printf("Ошибка Marshal: %v", err)
		}
		message := kafka.Message{
			Value: jsonOrder,
		}
		for {
			err = writer.WriteMessages(ctx, message)
			if err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
				time.Sleep(5 * time.Second)
			} else {
				log.Printf("Отправлен заказ UID=%v", order.OrderUID)
				break
			}
		}
	}
}
