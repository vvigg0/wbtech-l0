package service

import (
	"database/sql"
	"encoding/json"
	"l0/internal/models"
	"log"
	"os"
)

func GetFromDB(orderId string) ([]byte, error) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Ошибка открытия БД: %v", err)
	}
	defer db.Close()
	var order models.Order
	err = db.QueryRow("SELECT * FROM orders "+
		"WHERE order_uid=$1", orderId).Scan(
		&order.OrderUID, &order.TrackNumber,
		&order.Entry, &order.Locale,
		&order.InternalSign, &order.CustomerId,
		&order.DeliveryService, &order.ShardKey,
		&order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		return []byte{}, err
	}
	err = db.QueryRow("SELECT name,phone,"+
		"zip,city,address,region,"+
		"email FROM deliveries "+
		"WHERE order_uid=$1", orderId).Scan(
		&order.Delivery.Name, &order.Delivery.Phone,
		&order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email)
	if err != nil {
		return []byte{}, err
	}
	err = db.QueryRow("SELECT transaction,request_id,"+
		"currency,provider,amount,payment_dt,"+
		"bank,delivery_cost,goods_total,custom_fee "+
		"FROM payments WHERE order_uid=$1", orderId).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID,
		&order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDT,
		&order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return []byte{}, err
	}
	rows, err := db.Query("SELECT chrt_id,track_number,"+
		"price,rid,name,sale,size,"+
		"total_price,nm_id,brand,status "+
		"FROM items WHERE order_uid=$1", orderId)
	if err != nil {
		return []byte{}, err
	}
	defer rows.Close()
	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price,
			&item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
		if err != nil {
			return []byte{}, err
		}
		items = append(items, item)
	}
	order.Items = items
	jsonOrder, err := json.MarshalIndent(order, "", " ")
	if err != nil {
		log.Printf("Ошибка Marshal: %v", err)
		return []byte{}, err
	}
	return jsonOrder, nil
}
func InsertToDB(order models.Order) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()
	if err != nil {
		log.Println("Ошибка открытия базы данных: ", err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Println("Ошибка в транзакции: ", err)
	}
	_, err = tx.Exec("INSERT INTO orders(order_uid,track_number,entry,locale,internal_signature,customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)",
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSign, order.CustomerId, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	for _, item := range order.Items {
		_, err := tx.Exec("INSERT INTO items (order_uid,chrt_id,track_number,price,rid,name, sale,size,total_price, nm_id,brand,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)",
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}
	_, err = tx.Exec("INSERT INTO payments(order_uid,transaction,request_id,currency,provider,amount,payment_dt,bank,delivery_cost,goods_total,custom_fee) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)",
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	_, err = tx.Exec("INSERT INTO deliveries(order_uid,name,phone,zip,city,address,region,email) VALUES($1,$2,$3,$4,$5,$6,$7,$8)",
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}
