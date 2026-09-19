package main

func HandleMessage(db *DB, topic string, value []byte) error {
	orderID := extractOrderID(value)

	switch topic {
	case "payment-success":
		return db.UpdateOrderStatus(orderID, "CONFIRMED", "")
	case "payment-failed":
		return db.UpdateOrderStatus(orderID, "CANCELLED", "Payment failed")
	case "inventory-failed":
		return db.UpdateOrderStatus(orderID, "CANCELLED", "Insufficient inventory")
	}

	return nil
}
