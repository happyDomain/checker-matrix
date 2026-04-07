// Package checker implements the Matrix federation checker for happyDomain.
//
// It queries a Matrix Federation Tester instance (by default the public one
// hosted at https://federationtester.matrix.org) to verify that a Matrix
// homeserver is correctly federating, then exposes the result both as an
// observation and as a rich HTML report.
package checker

import (
	sdk "git.happydns.org/checker-sdk-go/checker"
)

// ObservationKeyMatrix is the observation key for Matrix federation test data.
const ObservationKeyMatrix sdk.ObservationKey = "matrix_federation"

// MatrixFederationData is the full payload returned by the Matrix Federation
// Tester API and stored as the observation.
type MatrixFederationData struct {
	WellKnownResult struct {
		Server         string `json:"m.server"`
		Result         string `json:"result"`
		CacheExpiresAt int64  `json:"CacheExpiresAt"`
	} `json:"WellKnownResult"`
	DNSResult struct {
		SRVSkipped bool   `json:"SRVSkipped"`
		SRVCName   string `json:"SRVCName"`
		SRVRecords []struct {
			Target   string `json:"Target"`
			Port     uint16 `json:"Port"`
			Priority uint16 `json:"Priority"`
			Weight   uint16 `json:"Weight"`
		} `json:"SRVRecords"`
		SRVError *struct {
			Message string `json:"Message"`
		} `json:"SRVError"`
		Hosts map[string]struct {
			CName string   `json:"CName"`
			Addrs []string `json:"Addrs"`
		} `json:"Hosts"`
		Addrs []string `json:"Addrs"`
	} `json:"DNSResult"`
	ConnectionReports map[string]struct {
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
	} `json:"ConnectionReports"`
	ConnectionErrors map[string]struct {
		Message string `json:"Message"`
	} `json:"ConnectionErrors"`
	Version struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Error   string `json:"error,omitempty"`
	} `json:"Version"`
	FederationOK bool `json:"FederationOK"`
}
