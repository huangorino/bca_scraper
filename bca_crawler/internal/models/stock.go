package models

type BursaStockResponse struct {
	RecordsTotal    string          `json:"recordsTotal"`
	RecordsFiltered string          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
	Board           interface{}     `json:"board"`
}

type Stock struct {
	ID        int    `json:"id"`
	StockCode string `json:"stock_code"`
	Name      string `json:"stock_name"`
	Market    string `json:"market"`
	Status    string `json:"status"`
}

type StockRow struct {
	Index         int
	Name          string
	Code          string
	Market        string
	LastPrice     string
	ChangePrice   string
	ChangeValue   string
	ChangePercent string
	Volume        string
	Value         string
	Bid           string
	Ask           string
	BidVolume     string
	High          string
	Low           string
	Misc          string
}
