package main

import (
	"context"
	"database/sql"
	"l0/internal/cache"
	"l0/internal/handlers"
	myKafka "l0/internal/kafka"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}
	cache := cache.NewTTLMap()
	err = cache.Restore(db) //восстановление кэша по последним 3 записям(отсортированным по дате создания)
	if err != nil {
		log.Printf("Неудачная попытка восстановления кэша: %v", err)
	} else {
		log.Printf("Кэш успешно восстановлен!")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		go myKafka.GetOrder(ctx, db, cache) //получение сообщений из кафки
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/app/web/index.html") // возвращает html
	})
	mux.HandleFunc("/api/order/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleOrder(db, cache, w, r) // возвращает сырой json
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Не удалось запустить сервер: %v", err)
		}
	}()
	log.Println("order-service запущен")

	<-ctx.Done()
	log.Println("Получен сигнал завершения, начинаю graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке HTTP сервера: %v", err)
	}

	wg.Wait()

	if err := db.Close(); err != nil {
		log.Printf("Ошибка при закрытии БД: %v", err)
	}

	log.Println("graceful shutdown завершен")
}
