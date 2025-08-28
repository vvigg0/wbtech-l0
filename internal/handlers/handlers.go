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
	if !ok {
		jsonOrder, err := service.GetFromDB(db, orderId)
		if err != nil {
			log.Printf("Ошибка при запросе данных из БД: %v", err)
			fmt.Fprintf(w, "Ошибка при запросе данных из БД")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonOrder))
		log.Printf("Достали данные из БД за %v", time.Since(start))
	} else {
		jsonOrder, err := json.MarshalIndent(data, "", " ")
		if err != nil {
			log.Printf("Ошибка Marshal из кэша: %v", err)
			fmt.Fprintf(w, "Ошибка Marshal из кэша")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonOrder))
		log.Printf("Достали данные из кэша за %v", time.Since(start))
	}
}
