package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"l0/internal/cache"
	"l0/internal/service"
	"log"
	"net/http"
	"strings"
	"time"
)

func HandleOrder(db *sql.DB, cache *cache.TTLMap, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	splitUrl := strings.Split(r.URL.String(), "/")
	orderId := splitUrl[len(splitUrl)-1]
	data, ok := cache.Get(orderId)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if !ok {
		jsonOrder, err := service.GetFromDB(db, orderId)
		if err != nil {
			log.Printf("Ошибка при запросе данных из БД: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Ошибка при запросе данных из БД")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonOrder))
		log.Printf("Достали данные из БД за %v", time.Since(start))
	} else {
		jsonOrder, err := json.MarshalIndent(data, "", " ")
		if err != nil {
			log.Printf("Ошибка Marshal из кэша: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Ошибка Marshal из кэша")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonOrder))
		log.Printf("Достали данные из кэша за %v", time.Since(start))
	}
}
