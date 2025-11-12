package myKafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"l0/internal/cache"
	"l0/internal/models"
	"l0/internal/service"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

var kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
	Brokers: []string{os.Getenv("KAFKA_URL")},
	Topic:   "orders",
})

func GetOrder(ctx context.Context, db *sql.DB, cache *cache.TTLMap) {
	var order models.Order
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_URL")},
		Topic:   "orders",
		GroupID: "order-service-consumer",
	})
	defer reader.Close()
	for {
		if ctx.Err() != nil {
			log.Println("GetOrder: контекст отменён, выходим")
			return
		}

		message, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("GetOrder: чтение прервано контекстом, выходим")
				return
			}
			log.Println("Ошибка при прочтении сообщения: ", err)
			continue
		}

		err = json.Unmarshal(message.Value, &order)
		if err != nil {
			log.Println("Не удалось распарсить JSON: ", err)
			// чтобы не зациклиться
			err = reader.CommitMessages(ctx, message)
			if err != nil {
				log.Println("ошибка коммита:", err)
			}
			continue
		}

		if err := service.ValidateOrder(order); err != nil {
			log.Printf("Невалидный заказ, пропускаем: %v", err)
			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Println("Ошибка коммита(невалидный заказ):", err)
			}
			continue
		}

		err = service.InsertToDB(db, order)
		if err != nil {
			log.Printf("Ошибка вставки в БД: %v", err)
			continue
		}

		cache.Set(order.OrderUID, order)

		err = reader.CommitMessages(ctx, message)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("GetOrder: контекст отменён во время коммита, выходим")
			}
			log.Println("ошибка коммита:", err)
		}
		log.Printf("Получен заказ: %v", order.OrderUID)
	}
}
func SendOrder(orders int, ch chan models.Order) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	writer := kafkaWriter
	defer writer.Close()
	for i := 0; i < orders; i++ {
		order := <-ch
		jsonOrder, err := json.Marshal(order)
		if err != nil {
			log.Printf("Ошибка Marshal: %v", err)
			continue
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
