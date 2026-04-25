package checker

import (
	"context"
	"fmt"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// srvRecordsRule checks _matrix-fed._tcp / _matrix._tcp SRV delegation: was
// the lookup successful, and does it yield at least one record (or was it
// legitimately skipped because of a CNAME/well-known path)?
type srvRecordsRule struct{}

func (r *srvRecordsRule) Name() string { return "matrix.srv_records" }
func (r *srvRecordsRule) Description() string {
	return "Checks that the Matrix SRV lookup (_matrix-fed._tcp / _matrix._tcp) succeeded or was legitimately skipped."
}

func (r *srvRecordsRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, _ sdk.CheckerOptions) []sdk.CheckState {
	data, errSt := loadMatrixData(ctx, obs)
	if errSt != nil {
		return []sdk.CheckState{*errSt}
	}

	dns := data.DNSResult

	if dns.SRVError != nil {
		return []sdk.CheckState{critState("matrix.srv_records.error", fmt.Sprintf("SRV lookup error: %s", dns.SRVError.Message))}
	}

	if dns.SRVSkipped {
		msg := "SRV lookup skipped by the federation tester."
		if dns.SRVCName != "" {
			msg = fmt.Sprintf("SRV lookup skipped (CNAME: %s).", dns.SRVCName)
		}
		return []sdk.CheckState{unknownState("matrix.srv_records.skipped", msg)}
	}

	if len(dns.SRVRecords) == 0 {
		return []sdk.CheckState{infoState(
			"matrix.srv_records.absent",
			"No Matrix SRV records published (federation may still work via well-known).",
		)}
	}

	return []sdk.CheckState{passState("matrix.srv_records.ok", fmt.Sprintf("%d SRV record(s) published.", len(dns.SRVRecords)))}
}
