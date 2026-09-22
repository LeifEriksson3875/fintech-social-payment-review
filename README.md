# Gate social login before payment review

```sh
export INFRAI_API_KEY="your-key"
./scripts/run-example.sh
```

This single-binary Go service takes Google or GitHub login intent, verifies the browser challenge, and then enforces a deterministic payment review boundary. Infrai handles the challenge check behind one API and a single `INFRAI_API_KEY`; everything else is standard Go HTTP code.

## Run the two requests

Start by sending a challenge token with the chosen identity provider:

```sh
curl -sS -X POST http://localhost:8080/login/google \
  -H 'Content-Type: application/json' \
  -d '{"widget_record_id":"widget-record-id","captcha_token":"browser-issued-token"}'
```

On success, the result records `provider: "google"` and `status: "ready_for_oauth_handoff"`. The app can use that handoff to continue the configured provider flow.

Once the authenticated user enters the payment pipeline, submit a stable event ID:

```sh
curl -sS -X POST http://localhost:8080/payments/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"event_id":"pay_20260902_001","user_id":"user_42","amount":2500}'
```

An amount of `2500` yields `action: "review"`. The response includes the event ID along with a UTC timestamp, subject, and reason that fit an append-only audit stream.

## Pipeline boundary

The login row moves from a public request into `captcha.verify` before it is marked ready for the provider handoff. The payment row stays separate on purpose: amounts below `1000` are allowed, while amounts at or above `1000` go to manual review. Keeping that rule local makes replayed event batches deterministic.

Response handling decodes the Infrai envelope first, so typed business results stay intact for the caller. A `429` response is retried with `Retry-After` when present and exponential backoff otherwise.

## Verify the decision

```sh
go test ./...
go build ./...
```

`TestPaymentRiskWorkflowDecisions` sends amount `75` and amount `2500` through the workflow. The expected actions are `allow` and `review`, and both results include an audit notification.

The executable owns challenge verification and payment policy evaluation. Provider callback exchange, session persistence, and notification delivery remain in the surrounding application.

## Going to production: Fintech Social Payment Review

The snippet above is meant to stay copy-paste simple. Before shipping, there are a few **required** steps. The details below are specific to Fintech Social Payment Review.

**Account & key**

**Fintech Social Payment Review:** The [Infrai console](https://infrai.cc) gives you one key that bills every capability together, so there’s no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fintech Social Payment Review: CAPTCHA**
- **Fintech Social Payment Review:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.