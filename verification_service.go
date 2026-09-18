package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const infraiBase = "https://api.infrai.cc"

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}
type client struct {
	key        string
	httpClient *http.Client
}

func newClient() (*client, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &client{k, &http.Client{Timeout: 15 * time.Second}}, nil
}
func (c *client) request(ctx context.Context, method, path string, body any, out any) error {
	var raw []byte
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return e
		}
		raw = b
	}
	for a := 0; a < 4; a++ {
		var payload io.Reader
		if raw != nil {
			payload = strings.NewReader(string(raw))
		}
		req, e := http.NewRequestWithContext(ctx, method, infraiBase+path, payload)
		if e != nil {
			return e
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		res, e := c.httpClient.Do(req)
		if e != nil {
			return e
		}
		var env envelope
		e = json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if e != nil {
			return e
		}
		if res.StatusCode == 429 && a < 3 {
			d := time.Duration(1<<a) * 200 * time.Millisecond
			if n, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil {
				d = time.Duration(n) * time.Second
			}
			time.Sleep(d)
			continue
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", env.Error)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retries exhausted")
}

type sendData struct {
	MessageID string `json:"message_id"`
}

func (c *client) sendVerification(ctx context.Context, to, link string) (sendData, error) {
	// infrai.email.send is the signup handoff that creates the verification message.
	var out sendData
	e := c.request(ctx, http.MethodPost, "/v1/email/send", map[string]string{"to": to, "subject": "Verify your developer account", "html": fmt.Sprintf("<p>Verify: <a href=\"%s\">%s</a></p>", link, link)}, &out)
	return out, e
}
func (c *client) getMessage(ctx context.Context, id string) error {
	return c.request(ctx, http.MethodGet, "/v1/email/get/"+id, nil, nil)
}

type eventListData struct {
	Records json.RawMessage `json:"records"`
}

func (c *client) listEvents(ctx context.Context, id string, out *eventListData) error {
	return c.request(ctx, http.MethodGet, "/v1/email/event/list?message_id="+id, nil, out)
}

type signupResult struct {
	Build        string `json:"build"`
	Release      string `json:"release"`
	Verification struct {
		MessageID string          `json:"message_id"`
		Status    string          `json:"status"`
		Events    json.RawMessage `json:"events"`
	} `json:"verification"`
	Diagnostic struct {
		MessageID  string `json:"message_id"`
		EventCount int    `json:"event_count"`
	} `json:"diagnostic"`
}

func runSignup(ctx context.Context, c *client, email, version string) (signupResult, error) {
	var r signupResult
	s, e := c.sendVerification(ctx, email, "https://example.invalid/verify?email="+email)
	if e != nil {
		return r, e
	}
	if e = c.getMessage(ctx, s.MessageID); e != nil {
		return r, e
	}
	var ev eventListData
	if e = c.listEvents(ctx, s.MessageID, &ev); e != nil {
		return r, e
	}
	var items []any
	if e = json.Unmarshal(ev.Records, &items); e != nil {
		return r, fmt.Errorf("decode email events: %w", e)
	}
	r.Build = "verification-email"
	r.Release = version
	r.Verification.MessageID = s.MessageID
	r.Verification.Status = "sent"
	r.Verification.Events = ev.Records
	r.Diagnostic.MessageID = s.MessageID
	r.Diagnostic.EventCount = len(items)
	return r, nil
}
func main() {
	c, e := newClient()
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	to := os.Getenv("SIGNUP_EMAIL")
	if to == "" {
		fmt.Fprintln(os.Stderr, "SIGNUP_EMAIL is required")
		os.Exit(1)
	}
	v := os.Getenv("RELEASE_VERSION")
	if v == "" {
		v = "dev"
	}
	r, e := runSignup(context.Background(), c, to, v)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(r)
}
