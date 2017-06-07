package coredns

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

type Config struct {
	EtcdEndpoints string
	Zones         string
}

type dnsOp struct {
	zones map[string][]dnsprovider.Zone
}

const dnsProviderId = "coredns"

func (c *Config) newDNSOp() (*dnsOp, error) {
	var dnsProvider dnsprovider.Interface

	var file io.Reader
	var lines []string
	lines = append(lines, "etcd-endpoints = "+c.EtcdEndpoints)
	lines = append(lines, "zones = "+c.Zones)
	config := "[global]\n" + strings.Join(lines, "\n") + "\n"
	file = bytes.NewReader([]byte(config))

	dnsProvider, err := dnsprovider.GetDnsProvider(dnsProviderId, file)
	if err != nil {
		return nil, err
	}

	z, ok := dnsProvider.Zones()
	if !ok {
		return nil, fmt.Errorf("no zones found")
	}

	zones, err := z.List()
	if err != nil {
		return nil, err
	}

	allZoneMap := make(map[string][]dnsprovider.Zone)
	for _, zone := range zones {
		name := EnsureDotSuffix(zone.Name())
		allZoneMap[name] = append(allZoneMap[name], zone)
	}

	return &dnsOp{
		zones: allZoneMap,
	}, nil
}

func EnsureDotSuffix(s string) string {
	if !strings.HasSuffix(s, ".") {
		s = s + "."
	}
	return s
}
