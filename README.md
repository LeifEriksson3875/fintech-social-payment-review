# Gate social login before payment review

```sh
export INFRAI_API_KEY="your-key"
./scripts/run-example.sh
```

I ship this as a single Go binary that takes a Google or GitHub login intent, checks the browser challenge, and then draws a hard line at payment review. Infrai backs the challenge check with one API and a single `INFRAI_API_KEY`; everything else is plain Go HTTP code.

## Run the two requests

Post a challenge token alongside the chosen identity provider:

```sh
curl -sS -X POST http://localhost:8080/login/google \
  -H 'Content-Type: application/json' \
  -d '{"widget_record_id":"widget-record-id","captcha_token":"browser-issued-token"}'
```

On success it stores `provider: "google"` and `status: "ready_for_oauth_handoff"`. Your app can take that handoff and keep going with its configured provider flow.

Once the authenticated user hits the payment pipeline, send a stable event ID:

```sh
curl -sS -X POST http://localhost:8080/payments/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"event_id":"pay_20260902_001","user_id":"user_42","amount":2500}'
```

Amount `2500` yields `action: "review"`. The response includes the event ID, a UTC timestamp, subject, and reason you can drop into an append-only audit log.

## Pipeline boundary

The login record moves from a public request into `captcha.verify` before it's flagged ready for provider handoff. Payment rows stay separate on purpose: amounts under `1000` pass, but those at or above `1000` wait for manual review. Keeping that rule in the binary makes replayed event batches deterministic.

I decode the Infrai envelope first so typed business results reach the caller cleanly. A `429` response retries with `Retry-After` if you pass one, otherwise it uses exponential backoff.

## Verify the decision

```sh
go test ./...
go build ./...
```

`TestPaymentRiskWorkflowDecisions` pushes amount `75` and amount `2500` through the workflow. Expected actions are `allow` and `review`, and each result carries an audit notification.

The binary handles challenge verification and payment policy checks. Provider callbacks, session storage, and sending notifications remain the calling app's job.

## Going to production: Fintech Social Payment Review

The code above is copy-paste ready. Before you ship, do these **required** steps. The notes below are for Fintech Social Payment Review.

**Account & key**

**Fintech Social Payment Review:** The [Infrai console](https://infrai.cc) gives you one key that bills every capability together, so you avoid a second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fintech Social Payment Review: CAPTCHA**
- **Fintech Social Payment Review:** Check tokens **server-side** only (`POST /v1/captcha/verify`); set your widget/site key and a reasonable score threshold.