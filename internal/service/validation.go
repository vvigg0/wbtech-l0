package service

import (
	"fmt"
	"strings"
	"time"

	"l0/internal/models"
)

func ValidateOrder(o models.Order) error {
	var errs []string

	// ===== orders =====

	if isEmpty(o.OrderUID) {
		errs = append(errs, "order_uid is empty")
	}
	if isEmpty(o.TrackNumber) {
		errs = append(errs, "track_number is empty")
	}
	if isEmpty(o.Entry) {
		errs = append(errs, "entry is empty")
	}
	if isEmpty(o.Locale) {
		errs = append(errs, "locale is empty")
	}
	if isEmpty(o.InternalSign) {
		errs = append(errs, "internal_signature is empty")
	}
	if isEmpty(o.CustomerId) {
		errs = append(errs, "customer_id is empty")
	}
	if isEmpty(o.DeliveryService) {
		errs = append(errs, "delivery_service is empty")
	}
	if isEmpty(o.ShardKey) {
		errs = append(errs, "shardkey is empty")
	}
	if o.SmID <= 0 {
		errs = append(errs, "sm_id must be > 0")
	}
	if o.DateCreated.IsZero() {
		errs = append(errs, "date_created is zero")
	} else if o.DateCreated.After(time.Now().Add(24 * time.Hour)) {
		errs = append(errs, "date_created is in the future")
	}
	if isEmpty(o.OofShard) {
		errs = append(errs, "oof_shard is empty")
	}

	// ===== delivery =====

	if err := validateDelivery(o.Delivery); err != nil {
		errs = append(errs, err.Error())
	}

	// ===== payment =====

	if err := validatePayment(o.Payment); err != nil {
		errs = append(errs, err.Error())
	}

	// ===== items =====

	if err := validateItems(o.Items); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("invalid order %s: %s", o.OrderUID, strings.Join(errs, "; "))
}

func validateDelivery(d models.Delivery) error {
	var errs []string

	if isEmpty(d.Name) {
		errs = append(errs, "delivery.name is empty")
	}
	if isEmpty(d.City) {
		errs = append(errs, "delivery.city is empty")
	}
	if isEmpty(d.Address) {
		errs = append(errs, "delivery.address is empty")
	}

	if d.Email != "" {
		if !strings.Contains(d.Email, "@") {
			errs = append(errs, "delivery.email has no '@'")
		} else {
			parts := strings.Split(d.Email, "@")
			if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
				errs = append(errs, "delivery.email format is invalid")
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(errs, "; "))
}

func validatePayment(p models.Payment) error {
	var errs []string

	if isEmpty(p.Transaction) {
		errs = append(errs, "payment.transaction is empty")
	}
	if isEmpty(p.Currency) {
		errs = append(errs, "payment.currency is empty")
	}
	if isEmpty(p.Provider) {
		errs = append(errs, "payment.provider is empty")
	}
	if p.Amount <= 0 {
		errs = append(errs, "payment.amount must be > 0")
	}
	if p.GoodsTotal <= 0 {
		errs = append(errs, "payment.goods_total must be > 0")
	}
	if p.DeliveryCost < 0 {
		errs = append(errs, "payment.delivery_cost must be >= 0")
	}
	if p.CustomFee < 0 {
		errs = append(errs, "payment.custom_fee must be >= 0")
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(errs, "; "))
}

func validateItems(items []models.Item) error {
	var errs []string

	if len(items) == 0 {
		return fmt.Errorf("items is empty")
	}

	for i, it := range items {
		prefix := fmt.Sprintf("items[%d]", i)

		if isEmpty(it.TrackNumber) {
			errs = append(errs, fmt.Sprintf("%s.track_number is empty", prefix))
		}
		if isEmpty(it.Name) {
			errs = append(errs, fmt.Sprintf("%s.name is empty", prefix))
		}
		if it.Price <= 0 {
			errs = append(errs, fmt.Sprintf("%s.price must be > 0", prefix))
		}
		if it.TotalPrice <= 0 {
			errs = append(errs, fmt.Sprintf("%s.total_price must be > 0", prefix))
		}
		if it.Sale < 0 {
			errs = append(errs, fmt.Sprintf("%s.sale must be >= 0", prefix))
		}
		if it.Status < 0 {
			errs = append(errs, fmt.Sprintf("%s.status must be >= 0", prefix))
		}
		if it.ChrtID <= 0 {
			errs = append(errs, fmt.Sprintf("%s.chrt_id must be > 0", prefix))
		}
		if it.NmId <= 0 {
			errs = append(errs, fmt.Sprintf("%s.nm_id must be > 0", prefix))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(errs, "; "))
}

// ===== helpers =====

func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
