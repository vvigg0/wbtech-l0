package main

import (
	myKafka "l0/internal/kafka"
	"l0/internal/models"
	"log"
	"math"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// ДЛЯ ИЗМЕНЕНИЯ КОЛИЧЕСТВА ЗАКАЗОВ ДЛЯ ГЕНЕРАЦИИ
const orders = 10

// Данные для создания рандомных заказов
var (
	providers       = []string{"wbpay", "yookassa", "stripe", "paypal"}
	currencies      = []string{"RUB", "USD", "EUR"}
	entries         = []string{"WBIL", "WBME", "WBFW"}
	locales         = []string{"ru", "en", "de"}
	deliveryService = []string{"meest", "cdek", "boxberry", "dpd"}
	shardKeys       = []string{"1", "3", "7", "9"}
	oofShards       = []string{"0", "1", "2"}
	sizes           = []string{"XS", "S", "M", "L", "XL"}
	statuses        = []int{201, 202, 203, 301, 302}
	brands          = []string{"Vivienne Sabo", "Maybelline", "L’Oreal", "Revlon", "Rimmel"}
)

func round(v float64) int {
	return int(math.Round(v))
}
func fakeItem(track string) models.Item {
	price := gofakeit.Number(200, 5000) // цена от 200 до 5000
	sale := gofakeit.Number(0, 60)      // скидка в процентах
	total := round(float64(price) * float64(100-sale) / 100.0)

	return models.Item{
		ChrtID:      gofakeit.Number(1_000_000, 9_999_999),
		TrackNumber: track,
		Price:       price,
		Rid:         gofakeit.HexUint(128), // псевдо rid
		Name:        gofakeit.ProductName(),
		Sale:        sale,
		Size:        gofakeit.RandomString(sizes),
		TotalPrice:  total,
		NmId:        gofakeit.Number(1_000_000, 9_999_999),
		Brand:       gofakeit.RandomString(brands),
		Status:      gofakeit.RandomInt(statuses),
	}
}

// доставка (получатель)
func fakeDelivery() models.Delivery {
	addr := gofakeit.Address()
	return models.Delivery{
		Name:    gofakeit.Name(),
		Phone:   gofakeit.Phone(),
		Zip:     addr.Zip,
		City:    addr.City,
		Address: addr.Address,
		Region:  addr.State,
		Email:   gofakeit.Email(),
	}
}

func fakePayment(orderUID string, goodsTotal int, created time.Time) models.Payment {
	deliveryCost := gofakeit.Number(0, 1500)
	customFee := gofakeit.Number(0, 100)
	amount := goodsTotal + deliveryCost + customFee
	payTime := created.Add(time.Duration(gofakeit.Number(0, 3)) * time.Hour)
	return models.Payment{
		Transaction:  orderUID, // пусть будет тот же uid
		RequestID:    gofakeit.UUID(),
		Currency:     gofakeit.RandomString(currencies),
		Provider:     gofakeit.RandomString(providers),
		Amount:       amount,
		PaymentDT:    payTime.Unix(),
		Bank:         gofakeit.Company(),
		DeliveryCost: deliveryCost,
		GoodsTotal:   goodsTotal,
		CustomFee:    customFee,
	}
}
func fakeOrder() models.Order {
	orderUID := gofakeit.UUID()
	track := gofakeit.LetterN(12)

	// товары: 1–3 шт.
	itemCount := gofakeit.Number(1, 3)
	items := make([]models.Item, 0, itemCount)
	sumGoods := 0
	for i := 0; i < itemCount; i++ {
		it := fakeItem(track)
		items = append(items, it)
		sumGoods += it.TotalPrice
	}

	//случайная дата создания
	from := time.Now().AddDate(-1, 0, 0)
	to := time.Now()
	created := gofakeit.DateRange(from, to)

	payment := fakePayment(orderUID, sumGoods, created)
	delivery := fakeDelivery()

	return models.Order{
		OrderUID:        orderUID,
		TrackNumber:     track,
		Entry:           gofakeit.RandomString(entries),
		Delivery:        delivery,
		Payment:         payment,
		Items:           items,
		Locale:          gofakeit.RandomString(locales),
		InternalSign:    gofakeit.HexUint(128),
		CustomerId:      gofakeit.Username(),
		DeliveryService: gofakeit.RandomString(deliveryService),
		ShardKey:        gofakeit.RandomString(shardKeys),
		SmID:            gofakeit.Number(1, 500),
		DateCreated:     created.UTC(),
		OofShard:        gofakeit.RandomString(oofShards),
	}
}
func genRandomOrder(n int, ch chan models.Order) {
	gofakeit.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		ch <- fakeOrder()
		time.Sleep(time.Duration(2 * time.Second))
	}
}
func main() {
	ch := make(chan models.Order)
	go genRandomOrder(orders, ch)
	log.Println("Создание сообщений для кафки...")
	myKafka.SendOrder(orders, ch)
}
