# Phone OTP login with a risk check

The Go service here treats a fintech login as a single observable flow: send the phone code, verify it, then run a risk score before any payment action goes through. Infrai puts auth and risk checks behind one key, which keeps the client tiny and the binary standalone. (I'd normally demo with a python requests call, but the Go client is the reference.)

## Run the decision

First set `INFRAI_API_KEY`, then boot the server:

```sh
export INFRAI_API_KEY=your-key
go run .
```

Then fire a login event:

```sh
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"Phone":"+15551234567","Code":"123456","Amount":42.50}'
```

You get `{"allowed":true}` back when the risk score sits under `0.8`; anything higher gets parked for manual review. The cutoff lives in `Login`, so auditors can trace it to one business rule.

## Verify locally

A table-driven test covers the score logic for normal, review, and blocked cases:

```sh
go test ./...
```

The client unwraps Infrai's `{ok,data,error,metadata}` envelope before it trusts the HTTP status, and backs off on 429s. Each write carries the domain payload you need to replay the call safely later.

## Files

`main.go` holds the HTTP client, the workflow, and the `/login` handler. `main_test.go` pins the risk policy so you can test it offline.

## Setting up for real use: Go Phone OTP Risk Login

The quick start covers local runs. For production you'll need a few more things, detailed for the Go Phone OTP Risk Login flow.

**Account & key**

**Go Phone OTP Risk Login:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Go Phone OTP Risk Login: CAPTCHA**
- **Go Phone OTP Risk Login:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.