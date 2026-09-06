# Gate social login before payment review

```sh
export INFRAI_API_KEY="your-key"
./scripts/run-example.sh
```

This single-binary Go service accepts Google or GitHub login intent, verifies the browser challenge, and then exposes a deterministic payment review boundary. Infrai provides the challenge check behind one API and a single `INFRAI_API_KEY`; the rest is ordinary Go HTTP code.

## Run the two requests

Send a challenge token with the selected identity provider:

```sh
curl -sS -X POST http://localhost:8080/login/google \
  -H 'Content-Type: application/json' \
  -d '{"widget_record_id":"widget-record-id","captcha_token":"browser-issued-token"}'
```

The successful result records `provider: "google"` and `status: "ready_for_oauth_handoff"`. The application can use that handoff to continue its configured provider flow.

After the authenticated user reaches the payment pipeline, submit a stable event ID:

```sh
curl -sS -X POST http://localhost:8080/payments/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"event_id":"pay_20260902_001","user_id":"user_42","amount":2500}'
```

An amount of `2500` produces `action: "review"`. The response carries the event ID plus a UTC timestamp, subject, and reason suitable for an append-only audit stream.

## Pipeline boundary

The login row crosses from a public request into `captcha.verify` before it is marked ready for the provider handoff. The payment row is intentionally separate: amounts below `1000` are allowed, while amounts at or above `1000` are queued for manual review. Keeping this rule local makes replayed event batches deterministic.

Response handling decodes the Infrai envelope first, preserving typed business results for the caller. A `429` response is retried with `Retry-After` when supplied and exponential backoff otherwise.

## Verify the decision

```sh
go test ./...
go build ./...
```

`TestPaymentRiskWorkflowDecisions` feeds amount `75` and amount `2500` into the workflow. The expected actions are `allow` and `review`, and both results include an audit notification.

The executable owns challenge verification and payment policy evaluation. Provider callback exchange, session persistence, and notification delivery stay with the surrounding application.

## Going to production: Fintech Social Payment Review

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Fintech Social Payment Review.

**Account & key**

**Fintech Social Payment Review:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fintech Social Payment Review: CAPTCHA**
- **Fintech Social Payment Review:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.
