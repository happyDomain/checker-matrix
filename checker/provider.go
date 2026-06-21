package checker

import (
	"net"
	"strconv"

	sdk "git.happydns.org/checker-sdk-go/checker"
	tlsct "git.happydns.org/checker-tls/contract"
)

// Provider returns a new matrix federation observation provider.
func Provider() sdk.ObservationProvider {
	return &matrixProvider{}
}

type matrixProvider struct{}

func (p *matrixProvider) Key() sdk.ObservationKey {
	return ObservationKeyMatrix
}

// DiscoverEntries publishes tls.endpoint.v1 entries for every Matrix federation
// endpoint the tester reached, plus any SRV targets the federation DNS exposes.
// This lets checker-tls independently monitor certificate health without relying
// on the external federation tester's schedule.
func (p *matrixProvider) DiscoverEntries(data any) ([]sdk.DiscoveryEntry, error) {
	d, ok := data.(*MatrixFederationData)
	if !ok || d == nil {
		return nil, nil
	}

	seen := map[string]struct{}{}
	var out []sdk.DiscoveryEntry

	add := func(host string, port uint16) error {
		key := host + ":" + strconv.Itoa(int(port))
		if _, dup := seen[key]; dup {
			return nil
		}
		seen[key] = struct{}{}
		entry, err := tlsct.NewEntry(tlsct.TLSEndpoint{
			Host: host,
			Port: port,
			SNI:  host,
		})
		if err != nil {
			return err
		}
		out = append(out, entry)
		return nil
	}

	addAddr := func(addr string) error {
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			// No port in address: use default Matrix federation port.
			host = addr
			portStr = "8448"
		}
		port64, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil || port64 == 0 {
			return nil
		}
		return add(host, uint16(port64))
	}

	// Endpoints the tester actually connected to are the primary source.
	for addr := range d.ConnectionReports {
		if err := addAddr(addr); err != nil {
			return nil, err
		}
	}
	// Include failed addresses too so checker-tls can retry on its own schedule.
	for addr := range d.ConnectionErrors {
		if err := addAddr(addr); err != nil {
			return nil, err
		}
	}

	// SRV records provide coverage when the tester itself was unreachable.
	for _, r := range d.DNSResult.SRVRecords {
		if r.Target == "" || r.Port == 0 {
			continue
		}
		if err := add(r.Target, r.Port); err != nil {
			return nil, err
		}
	}

	return out, nil
}
