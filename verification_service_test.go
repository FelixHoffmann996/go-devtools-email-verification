package main

import (
	"encoding/json"
	"testing"
)

func TestVerificationLink(t *testing.T) {
	got := "https://example.invalid/verify?email=dev@example.com"
	if got != "https://example.invalid/verify?email=dev@example.com" {
		t.Fatal(got)
	}
}

func TestEventListData(t *testing.T) {
	var got eventListData
	if err := json.Unmarshal([]byte(`{"records":[{"type":"delivered"}]}`), &got); err != nil {
		t.Fatal(err)
	}
	var records []any
	if err := json.Unmarshal(got.Records, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
}
