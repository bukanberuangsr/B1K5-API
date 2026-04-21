package models

type Customer struct {
	ID         int    `json:"id"`
	CustomerID string `json:"customer_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type CustomerProfile struct {
	ID         int    `json:"id"`
	CustomerID int    `json:"customer_id"` // Simpan ID (int) dari tabel Customer
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
}
