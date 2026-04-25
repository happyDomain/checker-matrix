package checker

import (
	"context"
	"fmt"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// Rules returns the full list of CheckRules exposed by the Matrix checker.
// Each rule covers a single concern so the UI can show a clear checklist
// rather than a single monolithic pass/fail line.
func Rules() []sdk.CheckRule {
	return []sdk.CheckRule{
		&federationOKRule{},
		&wellKnownRule{},
		&srvRecordsRule{},
		&connectionReachableRule{},
		&tlsChecksRule{},
		&versionRule{},
	}
}

// Rule returns the aggregate federation rule.
//
// Deprecated: prefer Rules() which exposes every concern individually. Kept
// for backward compatibility with callers that embed a single rule.
func Rule() sdk.CheckRule {
	return &federationOKRule{}
}

// loadMatrixData fetches the Matrix observation. On error returns a
// CheckState the caller should emit to short-circuit its rule.
func loadMatrixData(ctx context.Context, obs sdk.ObservationGetter) (*MatrixFederationData, *sdk.CheckState) {
	var data MatrixFederationData
	if err := obs.Get(ctx, ObservationKeyMatrix, &data); err != nil {
		return nil, &sdk.CheckState{
			Status:  sdk.StatusError,
			Message: fmt.Sprintf("Failed to get Matrix federation data: %v", err),
			Code:    "matrix.observation_error",
		}
	}
	return &data, nil
}

func passState(code, message string) sdk.CheckState {
	return sdk.CheckState{Status: sdk.StatusOK, Message: message, Code: code}
}

func infoState(code, message string) sdk.CheckState {
	return sdk.CheckState{Status: sdk.StatusInfo, Message: message, Code: code}
}

func warnState(code, message string) sdk.CheckState {
	return sdk.CheckState{Status: sdk.StatusWarn, Message: message, Code: code}
}

func critState(code, message string) sdk.CheckState {
	return sdk.CheckState{Status: sdk.StatusCrit, Message: message, Code: code}
}

func unknownState(code, message string) sdk.CheckState {
	return sdk.CheckState{Status: sdk.StatusUnknown, Message: message, Code: code}
}
