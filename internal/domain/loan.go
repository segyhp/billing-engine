package domain

import "time"

type Loan struct {
	ID            string
	Amount        float64 `json:"amount"`
	InterestRate  float64 `json:"interest_rate"`
	WeeklyPayment float64 `json:"weekly_payment"`
	Schedule      []PaymentSchedule
	Payments      []Payment
	CreatedAt     time.Time
}
