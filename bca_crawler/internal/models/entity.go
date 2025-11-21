package models

type Entity struct {
	ID          int    `json:"id,omitempty"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	StockCode   string `json:"stock_code,omitempty"`
	ICNumber    string `json:"ic_number,omitempty"`
	Age         int    `json:"age,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}
