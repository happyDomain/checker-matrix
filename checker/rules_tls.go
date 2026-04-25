package checker

import (
	"context"
	"fmt"
	"strings"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// tlsChecksRule reviews the TLS-level findings the federation tester
// reports for every endpoint it managed to reach: certificate validity,
// matching server name, future expiry, presence of an Ed25519 key, and so
// on. One CheckState is emitted per reachable endpoint so the UI can pin
// the outcome on the exact address.
type tlsChecksRule struct{}

func (r *tlsChecksRule) Name() string { return "matrix.tls_checks" }
func (r *tlsChecksRule) Description() string {
	return "Reviews the TLS posture on every reachable federation endpoint (certificate chain, hostname match, Ed25519 key, …)."
}

func (r *tlsChecksRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, _ sdk.CheckerOptions) []sdk.CheckState {
	data, errSt := loadMatrixData(ctx, obs)
	if errSt != nil {
		return []sdk.CheckState{*errSt}
	}

	if len(data.ConnectionReports) == 0 {
		return []sdk.CheckState{infoState("matrix.tls_checks.skipped", "No endpoint reached: TLS posture could not be assessed.")}
	}

	out := make([]sdk.CheckState, 0, len(data.ConnectionReports))
	for addr, cr := range data.ConnectionReports {
		var problems []string
		if !cr.Checks.MatchingServerName {
			problems = append(problems, "server name does not match certificate")
		}
		if !cr.Checks.FutureValidUntilTS {
			problems = append(problems, "certificate expired or near expiry")
		}
		if !cr.Checks.ValidCertificates {
			problems = append(problems, "certificate chain is invalid")
		}
		if !cr.Checks.HasEd25519Key {
			problems = append(problems, "no Ed25519 signing key advertised")
		}
		if !cr.Checks.AllEd25519ChecksOK {
			problems = append(problems, "Ed25519 key verification failed")
		}
		for _, e := range cr.Errors {
			if e != "" {
				problems = append(problems, e)
			}
		}

		if len(problems) == 0 && cr.Checks.AllChecksOK {
			st := passState("matrix.tls_checks.ok", "All TLS checks passed.")
			st.Subject = addr
			out = append(out, st)
			continue
		}

		msg := "TLS checks failed."
		if len(problems) > 0 {
			msg = fmt.Sprintf("TLS checks failed: %s.", strings.Join(problems, "; "))
		}
		st := critState("matrix.tls_checks.fail", msg)
		st.Subject = addr
		out = append(out, st)
	}
	return out
}
