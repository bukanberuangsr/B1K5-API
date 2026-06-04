package controllers

import "B1K5-API/internal/utils"

type customerIdentity struct {
	ID         int
	CustomerID string
}

func getCustomerByIDParam(id string) (customerIdentity, error) {
	var customer customerIdentity

	err := utils.DB.QueryRow(`
		SELECT id, customer_id
		FROM customers
		WHERE id::text = $1 OR customer_id = $1
	`, id).Scan(&customer.ID, &customer.CustomerID)

	return customer, err
}
