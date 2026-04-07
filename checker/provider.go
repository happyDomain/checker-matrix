package checker

import (
	sdk "git.happydns.org/checker-sdk-go/checker"
)

// Provider returns a new matrix federation observation provider.
func Provider() sdk.ObservationProvider {
	return &matrixProvider{}
}

type matrixProvider struct{}

func (p *matrixProvider) Key() sdk.ObservationKey {
	return ObservationKeyMatrix
}

// Definition implements sdk.CheckerDefinitionProvider.
func (p *matrixProvider) Definition() *sdk.CheckerDefinition {
	return Definition()
}
