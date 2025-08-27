package main

import (
	myKafka "l0/internal/kafka"
	"l0/internal/models"
	"math/rand"
	"strconv"
	"time"
)

// ДЛЯ ИЗМЕНЕНИЯ КОЛИЧЕСТВА ЗАКАЗОВ ДЛЯ ГЕНЕРАЦИИ
const startFrom = 5
const orders = 2

// Данные для создания рандомных заказов
// Order
var trackNumbers = []string{"TR123", "TR456", "TR789"}
var entries = []string{"WBIL", "WEB", "MOB"}
var locales = []string{"en", "ru", "de"}
var internalSigns = []string{"", "sig1", "sig2"}
var customerIDs = []string{"cust-1", "cust-2", "cust-3"}
var deliveryServices = []string{"meest", "dhl", "ups"}
var shardKeys = []string{"1", "2", "3"}
var smIDs = []int{99, 100, 101}
var oofShards = []string{"1", "2", "3"}

// Delivery
var deliveryNames = []string{"Ivan Petrov", "John Smith", "Anna Müller"}
var deliveryPhones = []string{"+79001234567", "+18005551234", "+4915123456789"}
var deliveryZips = []string{"123456", "90210", "10115"}
var deliveryCities = []string{"Moscow", "Los Angeles", "Berlin"}
var deliveryAddresses = []string{"Lenina 1", "Sunset Blvd 42", "Alexanderplatz 5"}
var deliveryRegions = []string{"Moscow Region", "California", "Berlin"}
var deliveryEmails = []string{"ivan@test.com", "john@test.com", "anna@test.com"}

// Payment
var requestIDs = []string{"req-1", "req-2", "req-3"}
var currencies = []string{"USD", "EUR", "RUB"}
var providers = []string{"wbpay", "paypal", "stripe"}
var amounts = []int{1000, 2000, 3000}
var paymentDTs = []int64{1637907727, 1638000000, 1638100000}
var banks = []string{"alpha", "sber", "tinkoff"}
var deliveryCosts = []int{150, 250, 350}
var goodsTotals = []int{500, 1000, 1500}
var customFees = []int{0, 10, 20}

// Items
var chrtIDs = []int{111, 222, 333}
var itemTrackNumbers = []string{"TR123", "TR456", "TR789"}
var prices = []int{100, 200, 300}
var rids = []string{"rid-aaa", "rid-bbb", "rid-ccc"}
var itemNames = []string{"T-shirt", "Shoes", "Hat"}
var sales = []int{10, 20, 30}
var sizes = []string{"S", "M", "L"}
var totalPrices = []int{90, 160, 270}
var nmIDs = []int{999, 888, 777}
var brands = []string{"Nike", "Adidas", "Puma"}
var statuses = []int{200, 201, 202}

var r *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomFrom[T any](list []T) T {
	return list[r.Intn(len(list))]
}
func genRandomOrder(n int, ch chan models.Order) {
	UID := startFrom
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		items := []models.Item{}
		for j := 0; j < 1+r.Intn(3); j++ {
			item := models.Item{
				ChrtID:      randomFrom(chrtIDs),
				TrackNumber: randomFrom(itemTrackNumbers),
				Price:       randomFrom(prices),
				Rid:         randomFrom(rids),
				Name:        randomFrom(itemNames),
				Sale:        randomFrom(sales),
				Size:        randomFrom(sizes),
				TotalPrice:  randomFrom(totalPrices),
				NmId:        randomFrom(nmIDs),
				Brand:       randomFrom(brands),
				Status:      randomFrom(statuses),
			}
			items = append(items, item)
		}
		payment := models.Payment{
			Transaction:  strconv.Itoa(UID) + "test",
			RequestID:    randomFrom(requestIDs),
			Currency:     randomFrom(currencies),
			Provider:     randomFrom(providers),
			Amount:       randomFrom(amounts),
			PaymentDT:    randomFrom(paymentDTs),
			Bank:         randomFrom(banks),
			DeliveryCost: randomFrom(deliveryCosts),
			GoodsTotal:   randomFrom(goodsTotals),
			CustomFee:    randomFrom(customFees),
		}
		delivery := models.Delivery{
			Name:    randomFrom(deliveryNames),
			Phone:   randomFrom(deliveryPhones),
			Zip:     randomFrom(deliveryZips),
			City:    randomFrom(deliveryCities),
			Address: randomFrom(deliveryAddresses),
			Region:  randomFrom(deliveryRegions),
			Email:   randomFrom(deliveryEmails),
		}
		order := models.Order{
			OrderUID:        strconv.Itoa(UID),
			TrackNumber:     randomFrom(trackNumbers),
			Entry:           randomFrom(entries),
			Delivery:        delivery,
			Payment:         payment,
			Items:           items,
			Locale:          randomFrom(locales),
			InternalSign:    randomFrom(internalSigns),
			CustomerId:      randomFrom(customerIDs),
			DeliveryService: randomFrom(deliveryServices),
			ShardKey:        randomFrom(shardKeys),
			SmID:            randomFrom(smIDs),
			DateCreated:     time.Now(),
			OofShard:        randomFrom(oofShards),
		}
		ch <- order
		UID++
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	}
}
func main() {
	ch := make(chan models.Order)
	go genRandomOrder(orders, ch)
	myKafka.SendOrder(orders, ch)
}
