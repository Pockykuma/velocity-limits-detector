package models

type AcceptedLoad struct {
	ID         uint   `json:"id" gorm:"primary_key"`
	CustomerID uint   `json:"customer_id"`
	LoadAmount string `json:"load_amount"`
}
