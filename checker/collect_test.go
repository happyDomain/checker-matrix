package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

func TestCollectMissingDomain(t *testing.T) {
	p := &matrixProvider{}
	if _, err := p.Collect(context.Background(), sdk.CheckerOptions{}); err == nil {
		t.Fatal("expected error when serviceDomain is empty")
	}
}

func TestCollectSuccess(t *testing.T) {
	const body = `{
		"WellKnownResult": {"m.server": "matrix.example.org:8448", "result": ""},
		"DNSResult": {"SRVSkipped": false, "SRVRecords": [{"Target": "matrix.example.org.", "Port": 8448, "Priority": 10, "Weight": 5}]},
		"ConnectionReports": {"1.2.3.4:8448": {"Checks": {"AllChecksOK": true, "MatchingServerName": true, "FutureValidUntilTS": true, "HasEd25519Key": true, "AllEd25519ChecksOK": true, "ValidCertificates": true}}},
		"ConnectionErrors": {},
		"Version": {"name": "Synapse", "version": "1.100.0"},
		"FederationOK": true
	}`

	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()

	p := &matrixProvider{}
	out, err := p.Collect(context.Background(), sdk.CheckerOptions{
		"serviceDomain":          "example.org.",
		"federationTesterServer": srv.URL + "/api/report?server_name=%s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotURL, "server_name=example.org") {
		t.Errorf("unexpected URL %q", gotURL)
	}
	data, ok := out.(*MatrixFederationData)
	if !ok || data == nil {
		t.Fatalf("expected *MatrixFederationData, got %T", out)
	}
	if !data.FederationOK {
		t.Error("expected FederationOK=true")
	}
	if data.Version.Name != "Synapse" {
		t.Errorf("unexpected version name %q", data.Version.Name)
	}
}

func TestCollectNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	p := &matrixProvider{}
	_, err := p.Collect(context.Background(), sdk.CheckerOptions{
		"serviceDomain":          "example.org",
		"federationTesterServer": srv.URL + "/?s=%s",
	})
	if err == nil {
		t.Fatal("expected error on 502 response")
	}
}

func TestCollectMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	p := &matrixProvider{}
	_, err := p.Collect(context.Background(), sdk.CheckerOptions{
		"serviceDomain":          "example.org",
		"federationTesterServer": srv.URL + "/?s=%s",
	})
	if err == nil {
		t.Fatal("expected decode error")
	}
}
