package domain

type Payment struct {
	ID      string  `json:"id"`
	LoanID  string  `json:"loan_id"`
	Amount  float64 `json:"amount"`
	DueDate string  `json:"due_date"`
	Paid    bool    `json:"paid"`
}

type PaymentSchedule struct {
	WeekNumber int     `json:"week_number"`
	AmountDue  float64 `json:"amount_due"`
	DueDate    string  `json:"due_date"`
}

func (Payment) TableName() string {
	return "payments"
}
