package workflow

import (
	"fmt"
	"time"
)

type PaymentEvent struct {
	EventID string  `json:"event_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
}

type PaymentDecision struct {
	EventID      string            `json:"event_id"`
	Action       string            `json:"action"`
	Notification AuditNotification `json:"notification"`
}

type AuditNotification struct {
	Kind      string    `json:"kind"`
	SubjectID string    `json:"subject_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentRiskWorkflow struct {
	now func() time.Time
}

func NewPaymentRiskWorkflow() *PaymentRiskWorkflow {
	return &PaymentRiskWorkflow{now: time.Now}
}

func (w *PaymentRiskWorkflow) Evaluate(event PaymentEvent) PaymentDecision {
	action, reason := decidePayment(event.Amount)
	return PaymentDecision{
		EventID: event.EventID,
		Action:  action,
		Notification: AuditNotification{
			Kind:      "payment_policy_decision",
			SubjectID: event.UserID,
			Reason:    reason,
			CreatedAt: w.now().UTC(),
		},
	}
}

func decidePayment(amount float64) (string, string) {
	if amount >= 1000 {
		return "review", fmt.Sprintf("amount %.2f reached manual review threshold", amount)
	}
	return "allow", fmt.Sprintf("amount %.2f is within automatic approval policy", amount)
}
