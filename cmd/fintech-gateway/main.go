package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/fintech-social-risk/internal/infrai"
	"example.com/fintech-social-risk/internal/workflow"
)

type server struct {
	client   *infrai.Client
	payments *workflow.PaymentRiskWorkflow
}

type loginRequest struct {
	WidgetRecordID string `json:"widget_record_id"`
	CaptchaToken   string `json:"captcha_token"`
}

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	s := &server{client: infrai.NewClient(apiKey), payments: workflow.NewPaymentRiskWorkflow()}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/{provider}", s.login)
	mux.HandleFunc("POST /payments/evaluate", s.evaluatePayment)
	log.Println("fintech gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	provider := strings.ToLower(r.PathValue("provider"))
	if provider != "google" && provider != "github" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider must be google or github"})
		return
	}
	var input loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || input.WidgetRecordID == "" || input.CaptchaToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "widget_record_id and captcha_token are required"})
		return
	}
	err := s.client.VerifyCaptcha(r.Context(), infrai.CaptchaVerifyRequest{
		WidgetRecordID: input.WidgetRecordID, Token: input.CaptchaToken,
		IP: r.RemoteAddr, Action: "social_login", ScoreThreshold: 0.7,
	})
	if err != nil {
		writeClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"provider": provider,
		"status":   "ready_for_oauth_handoff",
	})
}

func (s *server) evaluatePayment(w http.ResponseWriter, r *http.Request) {
	var event workflow.PaymentEvent
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payment event"})
		return
	}
	if event.EventID == "" || event.UserID == "" || event.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "event_id, user_id, and positive amount are required"})
		return
	}
	writeJSON(w, http.StatusOK, s.payments.Evaluate(event))
}

func writeClientError(w http.ResponseWriter, err error) {
	var apiErr *infrai.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Code, "message": apiErr.Message})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "dependency_error"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}
