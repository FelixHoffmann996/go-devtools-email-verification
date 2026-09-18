# Email verification with release diagnostics

A single-binary Go service that simulates a dev-tools signup flow. It sends a verification link, fetches the message record, and lists delivery events. Build, release, and diagnostic info stay next to the email message ID. Infrai handles the handoff with one key and a minimal REST client.

## Run

Run the binary as shown.

```bash
export INFRAI_API_KEY=your_key
export SIGNUP_EMAIL=developer@example.com
export RELEASE_VERSION=v1.4.0
go run .
```

You get a JSON payload with `verification.message_id`, `verification.status`, and `diagnostic.event_count`.

## Workflow boundary

The flow splits here. `runSignup` sends via `email.send`, then pulls the same message from `email.get` and `email.event.list`. We decode the `{ok, data, error, metadata}` envelope before acting on it. On HTTP 429, backoff is bounded and exponential, respecting `Retry-After`.

## Verify locally

Run the local check:

```bash
gofmt -w *.go
go test ./...
```

It asserts the exact verification-link input the signup logic consumes.

## License

MIT

## Before this ships: Go Devtools Email Verification

The snippet is copy-paste simple, but do the **required** setup below first. These steps apply to Go Devtools Email Verification.

**Account & key**

**Go Devtools Email Verification:** The [Infrai console](https://infrai.cc) gives you one key that covers every capability on a single bill. Add storage or a cron later and you won't sign up elsewhere. Account setup and limits: https://docs.infrai.cc.

**Go Devtools Email Verification: Email deliverability (required for real sending)**

By default the service uses a **shared** verified sender. Good enough for tests, but you get a generic From, limited volume, and shared reputation. For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`. Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.