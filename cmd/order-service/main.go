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
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}
	defer db.Close()
	cache := cache.NewTTLMap()
	err = cache.Restore(cache, db)
	if err != nil {
		log.Printf("Неудачная попытка восстановления кэша: %v", err)
	} else {
		log.Printf("Кэш успешно восстановлен!")
	}
	go myKafka.GetOrder(db, cache)
	http.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/app/web/index.html") // путь внутри контейнера
	})
	http.HandleFunc("/api/order/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleOrder(db, cache, w, r)
	})
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Не удалось запустить сервер: ", err)
	} else {
		log.Println("Сервер успешно запущен")
	}
}
