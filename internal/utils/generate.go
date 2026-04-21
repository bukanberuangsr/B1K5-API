package utils

import "fmt"

func GenerateCustomerID(id int) string {
	return fmt.Sprintf("CUS-%06d", id)
}
