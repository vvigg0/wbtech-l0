package main

import (
	"database/sql"
	"l0/internal/cache"
	"l0/internal/handlers"
	myKafka "l0/internal/kafka"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	cache := cache.NewTTLMap()
	go myKafka.GetOrder(cache)
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}
	db.Close()
	http.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) { handlers.HandleOrder(cache, w, r) })
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Не удалось запустить сервер: ", err)
	} else {
		log.Println("Сервер успешно запущен")
	}
}
