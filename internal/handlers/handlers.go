package handlers

import (
	"encoding/json"
	"fmt"
	"l0/internal/cache"
	"l0/internal/service"
	"log"
	"net/http"
	"strings"
	"time"
)

func HandleOrder(cache *cache.TTLMap, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	splitUrl := strings.Split(r.URL.String(), "/")
	orderId := splitUrl[len(splitUrl)-1]
	data, ok := cache.Get(orderId)
	if !ok {
		jsonOrder, err := service.GetFromDB(orderId)
		if err != nil {
			log.Printf("Ошибка при запросе данных из БД: %v", err)
		}
		fmt.Fprintf(w, string(jsonOrder))
		log.Printf("Достали данные из БД за %v", time.Since(start))
	} else {
		jsonOrder, err := json.MarshalIndent(data, "", " ")
		if err != nil {
			fmt.Fprintf(w, "Ошибка Marshal из кэша: %v", err)
		}
		fmt.Fprintf(w, string(jsonOrder))
		log.Printf("Достали данные из кэша за %v", time.Since(start))
	}

}
