package coredns

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"log"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	k8scoredns "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/coredns"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
)

type Config struct {
	EtcdEndpoints string
	Zones         string
}

type dnsOp struct {
	zones      map[string]dnsprovider.Zone
	changesets map[string]dnsprovider.ResourceRecordChangeset
}

type recordKey struct {
	RecordType string
	FQDN       string
}

type rrsetData struct {
	key   recordKey
	rdata []string
	ttl   int64
}

const dnsProviderId = "coredns"

func (c *Config) newDNSOp() (*dnsOp, error) {
	var file io.Reader
	var lines []string
	lines = append(lines, "etcd-endpoints = "+c.EtcdEndpoints)
	lines = append(lines, "zones = "+c.Zones)
	config := "[global]\n" + strings.Join(lines, "\n") + "\n"
	log.Printf("%v", config)
	file = bytes.NewReader([]byte(config))

	var provider dnsprovider.Interface

	if k8scoredns.ProviderName != dnsProviderId {
		return nil, fmt.Errorf("provider mismatch coreos != %s", dnsProviderId)
	}

	provider, err := dnsprovider.GetDnsProvider(dnsProviderId, file)

	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("unknown DNS provider %q", dnsProviderId)
	}

	z, ok := provider.Zones()
	if !ok {
		return nil, fmt.Errorf("no zones found")
	}

	zones, err := z.List()
	if err != nil {
		return nil, err
	}

	allZoneMap := make(map[string]dnsprovider.Zone)
	for _, zone := range zones {
		name := EnsureDotSuffix(zone.Name())
		allZoneMap[name] = zone
	}

	return &dnsOp{
		zones:      allZoneMap,
		changesets: make(map[string]dnsprovider.ResourceRecordChangeset),
	}, nil
}

func EnsureDotSuffix(s string) string {
	if !strings.HasSuffix(s, ".") {
		s = s + "."
	}
	return s
}

func (o *dnsOp) findZone(fqdn string) dnsprovider.Zone {
	zoneName := EnsureDotSuffix(fqdn)
	for {
		zone := o.zones[zoneName]
		if zone != nil {
			return zone
		}
		dot := strings.IndexByte(zoneName, '.')
		if dot == -1 {
			return nil
		}
		zoneName = zoneName[dot+1:]
	}
}

func (o *dnsOp) getChangeset(zone dnsprovider.Zone) (dnsprovider.ResourceRecordChangeset, error) {
	key := zone.Name() + "::" + zone.ID()
	changeset := o.changesets[key]
	if changeset == nil {
		rrsProvider, ok := zone.ResourceRecordSets()
		if !ok {
			return nil, fmt.Errorf("zone does not support resource records %q", zone.Name())
		}
		changeset = rrsProvider.StartChangeset()
		o.changesets[key] = changeset
	}

	return changeset, nil
}

func (o *dnsOp) getRecord(k recordKey) (dnsprovider.ResourceRecordSet, error) {
	fqdn := EnsureDotSuffix(k.FQDN)

	zone := o.findZone(fqdn)
	if zone == nil {
		return nil, fmt.Errorf("no suitable zone found for %q", fqdn)
	}

	rrsProvider, ok := zone.ResourceRecordSets()
	if !ok {
		return nil, fmt.Errorf("zone does not support resource records %q", zone.Name())
	}

	rrs, err := rrsProvider.Get(fqdn)
	if err != nil {
		return nil, fmt.Errorf("Failed to get DNS record %s with error: %v", fqdn, err)
	}
	for _, rr := range rrs {
		rrName := EnsureDotSuffix(rr.Name())
		if rrName != fqdn {
			log.Printf("Skipping delete of record %q (name != %s)", rrName, fqdn)
			continue
		}
		if string(rr.Type()) != string(k.RecordType) {
			log.Printf("Skipping delete of record %q (type %s != %s)", rrName, rr.Type(), k.RecordType)
			continue
		}

		log.Printf("Found resource record %s %s", rrName, rr.Type())
		return rr, nil
	}
	return nil, fmt.Errorf("Resource record %s %s not found", fqdn, k.RecordType)

}

func (o *dnsOp) deleteRecords(k recordKey) error {
	log.Printf("Deleting all records for %s", k)

	fqdn := EnsureDotSuffix(k.FQDN)

	zone := o.findZone(fqdn)
	if zone == nil {
		return fmt.Errorf("no suitable zone found for %q", fqdn)
	}

	rrsProvider, ok := zone.ResourceRecordSets()
	if !ok {
		return fmt.Errorf("zone does not support resource records %q", zone.Name())
	}

	rrs, err := rrsProvider.Get(fqdn)
	if err != nil {
		return fmt.Errorf("Failed to get DNS record %s with error: %v", fqdn, err)
	}

	cs, err := o.getChangeset(zone)
	if err != nil {
		return err
	}
	for _, rr := range rrs {
		rrName := EnsureDotSuffix(rr.Name())
		if rrName != fqdn {
			log.Printf("Skipping delete of record %q (name != %s)", rrName, fqdn)
			continue
		}
		if string(rr.Type()) != string(k.RecordType) {
			log.Printf("Skipping delete of record %q (type %s != %s)", rrName, rr.Type(), k.RecordType)
			continue
		}

		log.Printf("Deleting resource record %s %s", rrName, rr.Type())
		cs.Remove(rr)
	}
	return nil
}

func (o *dnsOp) updateRecords(k recordKey, newRecords []string, ttl int64) error {
	fqdn := EnsureDotSuffix(k.FQDN)

	zone := o.findZone(fqdn)
	if zone == nil {
		return fmt.Errorf("no suitable zone found for %q", fqdn)
	}

	rrsProvider, ok := zone.ResourceRecordSets()
	if !ok {
		return fmt.Errorf("zone does not support resource records %q", zone.Name())
	}

	var existing dnsprovider.ResourceRecordSet

	rrs, err := rrsProvider.Get(fqdn)
	if err != nil {
		return fmt.Errorf("error querying resource records for zone %q: %v", zone.Name(), err)
	}

	for _, rr := range rrs {
		rrName := EnsureDotSuffix(rr.Name())
		if rrName != fqdn {
			log.Printf("Skipping record %q (name != %s)", rrName, fqdn)
			continue
		}
		if string(rr.Type()) != string(k.RecordType) {
			log.Printf("Skipping record %q (type %s != %s)", rrName, rr.Type(), k.RecordType)
			continue
		}

		if existing != nil {
			log.Printf("Found multiple matching records: %v and %v", existing, rr)
		} else {
			log.Printf("Found matching record: %s %s", k.RecordType, rrName)
		}
		existing = rr
	}

	cs, err := o.getChangeset(zone)
	if err != nil {
		return err
	}

	if existing != nil {
		log.Printf("will replace existing dns record %s %s", existing.Type(), existing.Name())
		cs.Remove(existing)
	}

	log.Printf("Adding DNS changes to batch %s %s", k, newRecords)
	rr := rrsProvider.New(fqdn, newRecords, ttl, rrstype.RrsType(k.RecordType))
	cs.Add(rr)

	return nil
}

func (o *dnsOp) applyChangeset() error {
	var errors []error
	for key, changeset := range o.changesets {
		log.Printf("applying DNS changeset for zone %s", key)
		if err := changeset.Apply(); err != nil {
			log.Printf("error applying DNS changeset for zone %s: %v", key, err)
			errors = append(errors, fmt.Errorf("error applying DNS changeset for zone %s: %v", key, err))
		}
	}
	if len(errors) != 0 {
		return errors[0]
	}
	return nil
}
