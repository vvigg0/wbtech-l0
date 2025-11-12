package cache

import (
	"database/sql"
	"fmt"
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
	value, ok := t.m[key]
	t.l.RUnlock()
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

// Восстанавливает 3 записи из бд в кэш
// В случае ошибок по отдельным записям — просто логирует и пропускает их.
func (t *TTLMap) Restore(db *sql.DB) error {
	const q = `
SELECT
  o.order_uid, o.track_number, o.entry,
  d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
  p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
  o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard
FROM orders o
JOIN deliveries d ON o.order_uid = d.order_uid
JOIN payments   p ON o.order_uid = p.order_uid
ORDER BY o.date_created DESC
LIMIT 3`

	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("restore: query orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order

		if err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry,
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
			&order.DateCreated, &order.OofShard,
		); err != nil {
			log.Printf("restore: scan order '%s': %v — пропускаю", order.OrderUID, err)
			continue // пропускаем проблемный order
		}

		// достаём items по order_uid
		itemRows, err := db.Query(`
			SELECT chrt_id, track_number, price, rid, name, sale, size,
			       total_price, nm_id, brand, status
			FROM items
			WHERE order_uid = $1`, order.OrderUID)
		if err != nil {
			log.Printf("restore: items query for '%s': %v — пропускаю", order.OrderUID, err)
			continue // пропускаем заказ, если items не достали
		}

		var items []models.Item
		for itemRows.Next() {
			var it models.Item
			err := itemRows.Scan(
				&it.ChrtID, &it.TrackNumber, &it.Price,
				&it.Rid, &it.Name, &it.Sale, &it.Size,
				&it.TotalPrice, &it.NmId, &it.Brand, &it.Status)
			if err != nil {
				log.Printf("restore: scan item for '%s': %v — пропускаю весь заказ", order.OrderUID, err)
				// закрываем itemRows и пропускаем весь заказ, чтобы не класть в кэш полу-данные
				itemRows.Close()
				items = nil
				goto skipOrder
			}
			items = append(items, it)
		}
		if err := itemRows.Err(); err != nil {
			log.Printf("restore: items cursor for '%s': %v — пропускаю", order.OrderUID, err)
			itemRows.Close()
			continue
		}
		//закрываем после успешного обхода
		itemRows.Close()

		order.Items = items

		// Кладём в кэш
		t.Set(order.OrderUID, order)
		continue

	skipOrder:
		// сюда прыгаем, если сломался scan по items
		continue
	}

	// Проверяем ошибки внешнего rows
	if err := rows.Err(); err != nil {
		return fmt.Errorf("restore: orders cursor: %w", err)
	}

	return nil
}
