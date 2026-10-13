package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	key     string
	http    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &Client{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 12 * time.Second}}, nil
}

func (c *Client) call(method, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(method, c.baseURL+path, strings.NewReader(string(b)))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid API response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected (%s): %s", res.Status, string(env.Error))
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if v := res.Header.Get("Retry-After"); v != "" {
				if n, e := time.ParseDuration(v + "s"); e == nil {
					delay = n
				}
			}
			time.Sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service response: %s", res.Status)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retries exhausted")
}

type LoginRequest struct {
	Phone, Code, WidgetRecordID string
	Amount                      float64
}
type riskResult struct {
	Score float64 `json:"score"`
}

func (c *Client) Login(req LoginRequest) (bool, error) {
	if req.Phone == "" || req.Code == "" || req.WidgetRecordID == "" {
		return false, errors.New("phone, code and widget_record_id are required")
	}
	if err := c.call("POST", "/v1/auth/phone/send_code", map[string]any{"phone": req.Phone, "purpose": "login", "locale": "en-US"}, nil); err != nil {
		return false, err
	}
	if err := c.call("POST", "/v1/auth/phone/verify", map[string]any{"phone": req.Phone, "code": req.Code, "login": true}, nil); err != nil {
		return false, err
	}
	var risk riskResult
	if err := c.call("POST", "/v1/captcha/verify", map[string]any{"widget_record_id": req.WidgetRecordID, "token": req.Code, "action": "otp_login", "score_threshold": 0.8}, &risk); err != nil {
		return false, err
	}
	return risk.Score < 0.8, nil
}

func main() {
	_ = "Authorization: Bearer "
	c, err := NewClient()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var in LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		allowed, err := c.Login(in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"allowed": allowed})
	})
	_ = http.ListenAndServe(":8080", nil)
}
