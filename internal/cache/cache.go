package cache

import (
	"database/sql"
	"l0/internal/models"
	"log"
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type TTLMap struct {
	m map[string]*CacheItem
	l sync.RWMutex
}

func NewTTLMap() *TTLMap {
	return &TTLMap{
		m: make(map[string]*CacheItem),
	}
}

func (t *TTLMap) Set(key string, value interface{}) {
	t.l.Lock()
	defer t.l.Unlock()
	expiration := time.Now().Add(2 * time.Minute).Unix()
	t.m[key] = &CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

func (t *TTLMap) Get(key string) (interface{}, bool) {
	t.l.RLock()
	defer t.l.RUnlock()
	value, ok := t.m[key]
	if !ok {
		return nil, false
	}
	if time.Now().Unix() > value.Expiration {
		t.Delete(key)
		return nil, false
	}
	return value.Value, true
}

func (t *TTLMap) Delete(key string) {
	t.l.Lock()
	defer t.l.Unlock()
	delete(t.m, key)
}

func (t *TTLMap) Restore(cache *TTLMap, db *sql.DB) error {
	rows, err := db.Query(`
	SELECT o.order_uid,o.track_number,o.entry,
			d.name,d.phone,d.zip,d.city,d.address,d.region,d.email,
			p.transaction,p.request_id,p.currency,p.provider,p.amount,p.payment_dt,p.bank,p.delivery_cost,p.goods_total,p.custom_fee,
		   o.locale,o.internal_signature,o.customer_id,o.delivery_service,o.shardkey,o.sm_id,o.date_created,o.oof_shard
		FROM orders o
		JOIN deliveries d ON o.order_uid=d.order_uid
		JOIN payments p ON o.order_uid=p.order_uid
		ORDER BY date_created DESC
		LIMIT 3`)
	if err != nil {
		return err
	}
	var order models.Order
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry,
			&order.Delivery.Name, &order.Delivery.Phone,
			&order.Delivery.Zip, &order.Delivery.City,
			&order.Delivery.Address, &order.Delivery.Region,
			&order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.RequestID,
			&order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.PaymentDT,
			&order.Payment.Bank, &order.Payment.DeliveryCost,
			&order.Payment.GoodsTotal, &order.Payment.CustomFee,
			&order.Locale, &order.InternalSign, &order.CustomerId,
			&order.DeliveryService, &order.ShardKey, &order.SmID,
			&order.DateCreated, &order.OofShard)
		if err != nil {
			log.Printf("Ошибка скана в структуру Orders: %v", err)
		}
		for _, v := range order.OrderUID {
			rows, err := db.Query("SELECT chrt_id,track_number,"+
				"price,rid,name,sale,size,"+
				"total_price,nm_id,brand,status "+
				"FROM items WHERE order_uid=$1", v)
			if err != nil {
				log.Printf("Ошибка при запросе в items: %v", err)
				continue
			}
			defer rows.Close()
			var items []models.Item
			for rows.Next() {
				var item models.Item
				err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price,
					&item.Rid, &item.Name, &item.Sale, &item.Size,
					&item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
				if err != nil {
					return err
				}
				items = append(items, item)
			}
			order.Items = items
		}
		cache.Set(order.OrderUID, order)
	}
	return nil
}
