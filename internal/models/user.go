package models

type Customer struct {
	ID           int    `json:"id"`
	CustomerID   string `json:"customer_id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

type CustomerProfile struct {
	ID         int    `json:"id"`
	CustomerID int    `json:"customer_id"`
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	CreatedAt  string `json:"created_at"`
}

type Activity struct {
	ID           int    `json:"id"`
	CustomerID   int    `json:"customer_id"`
	ActivityType string `json:"activity_type"`
	Feature      string `json:"feature"`
	Metadata     string `json:"metadata"`
	CreatedAt    string `json:"created_at"`
}

type Segmentation struct {
	ID          int     `json:"id"`
	CustomerID  int     `json:"customer_id"`
	SegmentID   int     `json:"segment_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	CreatedAt   string  `json:"created_at"`
}
