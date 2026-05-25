package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/opusdns/opusdns-go-client/models"
	"github.com/opusdns/opusdns-go-client/opusdns"
)

const (
	// Version is the webhook version, used in User-Agent headers.
	Version = "1.0.0"

	// defaultTTL is the default TTL for DNS records.
	defaultTTL = 60

	// defaultSecretKey is the default key within a Kubernetes Secret.
	defaultSecretKey = "api-key"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName, &opusDNSSolver{})
}

// opusDNSSolver implements the DNS01 solver for OpusDNS.
type opusDNSSolver struct {
	kubeClient *kubernetes.Clientset
}

// opusDNSConfig holds the configuration for the OpusDNS solver.
type opusDNSConfig struct {
	// APIKeySecretRef references the Secret containing the OpusDNS API key.
	APIKeySecretRef corev1.SecretKeySelector `json:"apiKeySecretRef"`
	// APIEndpoint is the OpusDNS API endpoint (optional, defaults to https://api.opusdns.com).
	APIEndpoint string `json:"apiEndpoint,omitempty"`
	// TTL for DNS records (optional, defaults to 60).
	TTL int `json:"ttl,omitempty"`
	// ZoneName is the DNS zone name (optional, auto-detected if not provided).
	ZoneName string `json:"zoneName,omitempty"`
}

// Name returns the solver name.
func (s *opusDNSSolver) Name() string {
	return "opusdns"
}

// Present creates a TXT record for the ACME DNS01 challenge.
func (s *opusDNSSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	klog.V(2).Infof("Present: namespace=%s, zone=%s, fqdn=%s",
		ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)

	client, cfg, err := s.newClient(ch)
	if err != nil {
		return err
	}

	ctx := context.Background()
	fqdn := strings.TrimSuffix(ch.ResolvedFQDN, ".")

	zoneName, err := s.resolveZone(ctx, client, cfg, fqdn)
	if err != nil {
		return err
	}

	recordName := extractRecordName(fqdn, zoneName)
	ttl := effectiveTTL(cfg.TTL)

	klog.V(4).Infof("Creating TXT record: %s in zone %s with value %s", recordName, zoneName, ch.Key)

	err = client.DNS.UpsertRecord(ctx, zoneName, models.Record{
		Name:  recordName,
		Type:  models.RRSetTypeTXT,
		TTL:   ttl,
		RData: quoteTXTRData(ch.Key),
	})
	if err != nil {
		return fmt.Errorf("failed to create TXT record %s in zone %s: %w", recordName, zoneName, err)
	}

	klog.V(2).Infof("Successfully created TXT record for %s", fqdn)
	return nil
}

// CleanUp removes the TXT record created for the ACME DNS01 challenge.
func (s *opusDNSSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	klog.V(2).Infof("CleanUp: namespace=%s, zone=%s, fqdn=%s",
		ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)

	client, cfg, err := s.newClient(ch)
	if err != nil {
		return err
	}

	ctx := context.Background()
	fqdn := strings.TrimSuffix(ch.ResolvedFQDN, ".")

	zoneName, err := s.resolveZone(ctx, client, cfg, fqdn)
	if err != nil {
		return err
	}

	recordName := extractRecordName(fqdn, zoneName)
	ttl := effectiveTTL(cfg.TTL)

	klog.V(4).Infof("Deleting TXT record: %s in zone %s", recordName, zoneName)

	err = client.DNS.DeleteRecord(ctx, zoneName, models.Record{
		Name:  recordName,
		Type:  models.RRSetTypeTXT,
		TTL:   ttl,
		RData: quoteTXTRData(ch.Key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete TXT record %s in zone %s: %w", recordName, zoneName, err)
	}

	klog.V(2).Infof("Successfully deleted TXT record for %s", fqdn)
	return nil
}

// Initialize sets up the Kubernetes client.
func (s *opusDNSSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	klog.V(2).Infof("Initializing OpusDNS solver (version %s)", Version)

	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	s.kubeClient = cl
	return nil
}

// resolveZone determines the zone name from config or by querying the API.
func (s *opusDNSSolver) resolveZone(ctx context.Context, client *opusdns.Client, cfg opusDNSConfig, fqdn string) (string, error) {
	if cfg.ZoneName != "" {
		return cfg.ZoneName, nil
	}
	zoneName, err := s.findZone(ctx, client, fqdn)
	if err != nil {
		return "", fmt.Errorf("failed to find zone for %s: %w", fqdn, err)
	}
	return zoneName, nil
}

// newClient creates an OpusDNS client from the challenge configuration.
func (s *opusDNSSolver) newClient(ch *v1alpha1.ChallengeRequest) (*opusdns.Client, opusDNSConfig, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return nil, cfg, fmt.Errorf("failed to load config: %w", err)
	}

	apiKey, err := s.getAPIKey(cfg, ch.ResourceNamespace)
	if err != nil {
		return nil, cfg, fmt.Errorf("failed to get API key: %w", err)
	}

	opts := []opusdns.Option{
		opusdns.WithAPIKey(apiKey),
		opusdns.WithTTL(effectiveTTL(cfg.TTL)),
		opusdns.WithUserAgent(fmt.Sprintf("cert-manager-webhook-opusdns/%s", Version)),
	}
	if cfg.APIEndpoint != "" {
		opts = append(opts, opusdns.WithAPIEndpoint(cfg.APIEndpoint))
	}

	client, err := opusdns.NewClient(opts...)
	if err != nil {
		return nil, cfg, fmt.Errorf("failed to create OpusDNS client: %w", err)
	}

	return client, cfg, nil
}

// getAPIKey retrieves the API key from the referenced Secret.
func (s *opusDNSSolver) getAPIKey(cfg opusDNSConfig, namespace string) (string, error) {
	secretName := cfg.APIKeySecretRef.Name
	if secretName == "" {
		return "", fmt.Errorf("apiKeySecretRef.name is required")
	}

	secretKey := cfg.APIKeySecretRef.Key
	if secretKey == "" {
		secretKey = defaultSecretKey
	}

	secret, err := s.kubeClient.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s/%s: %w", namespace, secretName, err)
	}

	apiKey, ok := secret.Data[secretKey]
	if !ok {
		return "", fmt.Errorf("secret %s/%s does not contain key %q", namespace, secretName, secretKey)
	}

	return strings.TrimSpace(string(apiKey)), nil
}

// findZone finds the matching zone for the given FQDN by querying the API.
// It returns the longest matching zone (most specific).
func (s *opusDNSSolver) findZone(ctx context.Context, client *opusdns.Client, fqdn string) (string, error) {
	zones, err := client.DNS.ListZones(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}

	// Find the longest matching zone (most specific match first).
	// Zone names from the API include a trailing dot (e.g. "example.com."),
	// so we normalize before comparing.
	parts := strings.Split(fqdn, ".")
	for i := 0; i < len(parts)-1; i++ {
		candidate := strings.Join(parts[i:], ".")
		for _, zone := range zones {
			if strings.TrimSuffix(zone.Name, ".") == candidate {
				klog.V(4).Infof("Found matching zone: %s for fqdn: %s", zone.Name, fqdn)
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("no matching zone found for %s (checked %d zones)", fqdn, len(zones))
}

// extractRecordName extracts the record name relative to the zone.
// E.g., extractRecordName("_acme-challenge.www.example.com", "example.com") = "_acme-challenge.www"
func extractRecordName(fqdn, zone string) string {
	if fqdn == zone {
		return "@"
	}
	suffix := "." + zone
	if strings.HasSuffix(fqdn, suffix) {
		return strings.TrimSuffix(fqdn, suffix)
	}
	return fqdn
}

// effectiveTTL returns the TTL to use, defaulting to defaultTTL.
func effectiveTTL(ttl int) int {
	if ttl <= 0 {
		return defaultTTL
	}
	return ttl
}

// quoteTXTRData wraps a TXT record value in double quotes if not already
// quoted. The OpusDNS API (and PowerDNS) store TXT rdata in quoted form
// (e.g. "123d=="). Sending the value pre-quoted ensures that both UPSERT
// and REMOVE comparisons on the server side find an exact match.
func quoteTXTRData(value string) string {
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		return value
	}
	return `"` + value + `"`
}

// loadConfig decodes the solver configuration.
func loadConfig(cfgJSON *extapi.JSON) (opusDNSConfig, error) {
	cfg := opusDNSConfig{}
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %w", err)
	}
	return cfg, nil
}
