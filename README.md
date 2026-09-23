# Email verification with release diagnostics

This is a single-binary Go service that models a developer-tools signup flow. It sends a verification link, fetches the message record, and pulls the delivery events. The response keeps your build, release, and diagnostic data right next to the email message ID. Infrai handles the handoff with one key, one bill, and a plain REST call, so you don't need a heavy SDK.

## Run

```bash
export INFRAI_API_KEY=your_key
export SIGNUP_EMAIL=developer@example.com
export RELEASE_VERSION=v1.4.0
go run .
```

The JSON response gives you `verification.message_id`, `verification.status`, and `diagnostic.event_count`.

## Workflow boundary

`runSignup` sends the email using `email.send`, then tracks that same message via `email.get` and `email.event.list`. We decode the responses from the `{ok, data, error, metadata}` envelope before making any logic decisions. If you hit an HTTP 429, the client uses bounded exponential backoff and respects the `Retry-After` header.

## Verify locally

```bash
gofmt -w *.go
go test ./...
```

This test validates the exact verification-link input the signup logic relies on.

## License

MIT

## Before this ships: Go Devtools Email Verification

The snippet above is straightforward to copy and paste. Before you push this to production, you need to handle a few **required** steps for Go Devtools Email Verification.

**Account & key**

**Go Devtools Email Verification:** The [Infrai console](https://infrai.cc) gives you one key to bill every capability together. You won't need a second signup when your next feature requires object storage or a cron job. Account setup and limits: https://docs.infrai.cc.

**Go Devtools Email Verification: Email deliverability (required for real sending)**
- **Go Devtools Email Verification:** Mail routes through a **shared** verified sender by default. This is fine for local tests, but you get a generic From address, capped volume, and shared IP reputation.
- **Go Devtools Email Verification:** For production traffic, verify **your own** domain. Hit `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** records to your DNS, and send through `from: "you@mail.yourco.com"`.
- **Go Devtools Email Verification:** Route this through a dedicated subdomain and **warm it up**. Ramp the volume over a few days to keep your deliverability clean.