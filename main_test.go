package main

import (
	"crypto/rand"
	"math/big"
	"os"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
)

var (
	zone      = os.Getenv("TEST_ZONE_NAME")
	dnsServer = os.Getenv("TEST_DNS_SERVER")
	fqdn      string
)

func TestRunsSuite(t *testing.T) {
	if zone == "" {
		t.Skip("TEST_ZONE_NAME not set, skipping integration tests")
	}

	fqdn = randomString(20) + "." + zone

	opts := []acmetest.Option{
		acmetest.SetResolvedZone(zone),
		acmetest.SetResolvedFQDN(fqdn),
		acmetest.SetAllowAmbientCredentials(false),
		acmetest.SetManifestPath("testdata/opusdns"),
	}

	if dnsServer != "" {
		opts = append(opts, acmetest.SetDNSServer(dnsServer))
	}

	fixture := acmetest.NewFixture(&opusDNSSolver{}, opts...)

	fixture.RunBasic(t)
	fixture.RunExtended(t)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[num.Int64()]
	}
	return string(b)
}
