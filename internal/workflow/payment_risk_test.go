package workflow

import (
	"testing"
	"time"
)

func TestPaymentRiskWorkflowDecisions(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   string
	}{
		{name: "ordinary payment allowed", amount: 75, want: "allow"},
		{name: "large payment reviewed", amount: 2500, want: "review"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workflow := NewPaymentRiskWorkflow()
			workflow.now = func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }
			got := workflow.Evaluate(PaymentEvent{
				EventID: "pay_001", UserID: "user_001", Amount: tc.amount,
			})
			if got.Action != tc.want {
				t.Fatalf("action = %q, want %q", got.Action, tc.want)
			}
			if got.Notification.SubjectID != "user_001" || got.Notification.CreatedAt.IsZero() {
				t.Fatalf("notification is not audit-ready: %+v", got.Notification)
			}
		})
	}
}
