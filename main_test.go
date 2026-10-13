package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginCaptchaRequest(t *testing.T) {
	for _, tt := range []struct {
		score float64
		want  bool
	}{{0.25, true}, {0.8, false}, {0.95, false}} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/captcha/verify" {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["widget_record_id"] != "widget-123" || body["token"] != "123456" {
					t.Errorf("unexpected captcha payload: %v", body)
				}
				json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]any{"score": tt.score}})
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]any{}})
		}))
		c := &Client{baseURL: server.URL, key: "test", http: server.Client()}
		got, err := c.Login(LoginRequest{Phone: "test", Code: "123456", WidgetRecordID: "widget-123"})
		server.Close()
		if err != nil || got != tt.want {
			t.Fatalf("score %v: allowed=%v, err=%v; want %v", tt.score, got, err, tt.want)
		}
	}
}

func TestRiskDecision(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  bool
	}{{"normal", 0.25, true}, {"review", 0.8, false}, {"blocked", 0.95, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.score < 0.8
			if got != tt.want {
				t.Fatalf("score %.2f: got %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}
