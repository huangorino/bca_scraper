package models

type Announcement struct {
	ID          int      `json:"id,omitempty"`
	AnnID       int      `json:"ann_id"`
	Title       string   `json:"title,omitempty"`
	Link        string   `json:"link"`
	CompanyName string   `json:"company_name,omitempty"`
	StockName   string   `json:"stock_name,omitempty"`
	DatePosted  string   `json:"date_posted,omitempty"`
	Category    string   `json:"category,omitempty"`
	RefNumber   string   `json:"ref_number,omitempty"`
	Content     string   `json:"content,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}
