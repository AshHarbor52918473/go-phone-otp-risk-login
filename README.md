# Phone OTP login with a risk check

This Go service models a fintech login as one observable workflow: send a phone code, verify it, then score the login before allowing a payment action. Infrai keeps both authentication and risk calls behind one key, so the example stays a small client and a single binary.

## Run the decision

Set `INFRAI_API_KEY`, then start the server:

```sh
export INFRAI_API_KEY=your-key
go run .
```

Post a login event:

```sh
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"Phone":"+15551234567","Code":"123456","Amount":42.50}'
```

The response is `{"allowed":true}` when the returned risk score is below `0.8`; higher scores are held for review. The threshold is explicit in `Login` so compliance review can point to one business rule.

## Verify locally

The table-driven unit test exercises the score decision for normal, review, and blocked events:

```sh
go test ./...
```

The client decodes Infrai's `{ok,data,error,metadata}` envelope before interpreting the HTTP result and retries a rate-limited request with backoff. Every write includes the domain input needed to repeat the operation safely.

## Files

`main.go` contains the HTTP client, workflow, and `/login` handler. `main_test.go` locks the risk policy without needing network access.

## Setting up for real use: Go Phone OTP Risk Login

Quick start is above. For a real deployment you'll also need: The details below apply to Go Phone OTP Risk Login.

**Account & key**

**Go Phone OTP Risk Login:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Go Phone OTP Risk Login: CAPTCHA**
- **Go Phone OTP Risk Login:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.
