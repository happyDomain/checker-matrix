package checker

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

type stubObs struct {
	data *MatrixFederationData
	err  error
}

func (s stubObs) Get(_ context.Context, _ sdk.ObservationKey, dest any) error {
	if s.err != nil {
		return s.err
	}
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func (s stubObs) GetRelated(_ context.Context, _ sdk.ObservationKey) ([]sdk.RelatedObservation, error) {
	return nil, nil
}

func eval(t *testing.T, rule sdk.CheckRule, data *MatrixFederationData, opts sdk.CheckerOptions) []sdk.CheckState {
	t.Helper()
	return rule.Evaluate(context.Background(), stubObs{data: data}, opts)
}

func TestFederationOKRulePass(t *testing.T) {
	data := &MatrixFederationData{FederationOK: true}
	data.Version.Name = "Synapse"
	data.Version.Version = "1.100.0"
	got := eval(t, &federationOKRule{}, data, sdk.CheckerOptions{"serviceDomain": "example.org."})
	if len(got) != 1 || got[0].Status != sdk.StatusOK {
		t.Fatalf("expected single OK state, got %+v", got)
	}
	if !strings.Contains(got[0].Message, "Synapse 1.100.0") {
		t.Errorf("expected version in message, got %q", got[0].Message)
	}
}

func TestFederationOKRuleFailDeterministicOrder(t *testing.T) {
	data := &MatrixFederationData{}
	data.ConnectionErrors = map[string]struct {
		Message string `json:"Message"`
	}{
		"z.example:8448": {Message: "boom z"},
		"a.example:8448": {Message: "boom a"},
		"m.example:8448": {Message: "boom m"},
	}
	first := eval(t, &federationOKRule{}, data, nil)[0].Message
	for range 5 {
		if eval(t, &federationOKRule{}, data, nil)[0].Message != first {
			t.Fatal("federation_ok message not stable across runs")
		}
	}
	idxA := strings.Index(first, "a.example")
	idxM := strings.Index(first, "m.example")
	idxZ := strings.Index(first, "z.example")
	if !(idxA < idxM && idxM < idxZ) {
		t.Errorf("expected sorted order a<m<z, got %q", first)
	}
}

func TestConnectionReachableUnknownWhenNoEndpoints(t *testing.T) {
	got := eval(t, &connectionReachableRule{}, &MatrixFederationData{}, nil)
	if len(got) != 1 || got[0].Status != sdk.StatusUnknown {
		t.Fatalf("expected single Unknown state, got %+v", got)
	}
}

func TestConnectionReachableSortedFailures(t *testing.T) {
	data := &MatrixFederationData{}
	data.ConnectionErrors = map[string]struct {
		Message string `json:"Message"`
	}{
		"b:1": {Message: "b err"},
		"a:1": {Message: "a err"},
	}
	got := eval(t, &connectionReachableRule{}, data, nil)
	if len(got) != 2 {
		t.Fatalf("expected 2 states, got %d", len(got))
	}
	if got[0].Subject != "a:1" || got[1].Subject != "b:1" {
		t.Errorf("subjects not sorted: %q, %q", got[0].Subject, got[1].Subject)
	}
}

func TestSRVRecordsSkipped(t *testing.T) {
	data := &MatrixFederationData{}
	data.DNSResult.SRVSkipped = true
	data.DNSResult.SRVCName = "matrix.example.org."
	got := eval(t, &srvRecordsRule{}, data, nil)
	if len(got) != 1 || got[0].Status != sdk.StatusUnknown {
		t.Fatalf("expected Unknown, got %+v", got)
	}
}

func TestVersionRulePass(t *testing.T) {
	data := &MatrixFederationData{}
	data.Version.Name = "Dendrite"
	data.Version.Version = "0.13.0"
	got := eval(t, &versionRule{}, data, nil)
	if len(got) != 1 || got[0].Status != sdk.StatusOK {
		t.Fatalf("expected OK, got %+v", got)
	}
}

func TestVersionRuleError(t *testing.T) {
	data := &MatrixFederationData{}
	data.Version.Error = "connection refused"
	got := eval(t, &versionRule{}, data, nil)
	if len(got) != 1 || got[0].Status != sdk.StatusWarn {
		t.Fatalf("expected Warn, got %+v", got)
	}
}

func TestWellKnownAbsent(t *testing.T) {
	got := eval(t, &wellKnownRule{}, &MatrixFederationData{}, nil)
	if len(got) != 1 || got[0].Status != sdk.StatusInfo {
		t.Fatalf("expected Info, got %+v", got)
	}
}

func TestTLSChecksDedupAndSorted(t *testing.T) {
	data := &MatrixFederationData{}
	data.ConnectionReports = map[string]struct {
		Certificates []struct {
			SubjectCommonName string   `json:"SubjectCommonName"`
			IssuerCommonName  string   `json:"IssuerCommonName"`
			SHA256Fingerprint string   `json:"SHA256Fingerprint"`
			DNSNames          []string `json:"DNSNames"`
		} `json:"Certificates"`
		Cipher struct {
			Version     string `json:"Version"`
			CipherSuite string `json:"CipherSuite"`
		} `json:"Cipher"`
		Checks struct {
			AllChecksOK        bool `json:"AllChecksOK"`
			MatchingServerName bool `json:"MatchingServerName"`
			FutureValidUntilTS bool `json:"FutureValidUntilTS"`
			HasEd25519Key      bool `json:"HasEd25519Key"`
			AllEd25519ChecksOK bool `json:"AllEd25519ChecksOK"`
			ValidCertificates  bool `json:"ValidCertificates"`
		} `json:"Checks"`
		Errors []string `json:"Errors"`
	}{
		"b:8448": {Errors: []string{"server name does not match certificate"}},
		"a:8448": {Errors: []string{"server name does not match certificate", "server name does not match certificate"}},
	}

	got := eval(t, &tlsChecksRule{}, data, nil)
	if len(got) != 2 {
		t.Fatalf("expected 2 states, got %d", len(got))
	}
	if got[0].Subject != "a:8448" || got[1].Subject != "b:8448" {
		t.Errorf("subjects not sorted: %q, %q", got[0].Subject, got[1].Subject)
	}
	if strings.Count(got[0].Message, "server name does not match certificate") != 1 {
		t.Errorf("expected dedup, got %q", got[0].Message)
	}
}

func TestLoadMatrixDataObservationError(t *testing.T) {
	rule := &federationOKRule{}
	got := rule.Evaluate(context.Background(), stubObs{err: context.Canceled}, nil)
	if len(got) != 1 || got[0].Status != sdk.StatusError {
		t.Fatalf("expected Error state, got %+v", got)
	}
}
