package cloudflare

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/viper"
	"github.com/we11adam/uddns/provider"
	"github.com/we11adam/uddns/updater"
)

type Config struct {
	Email    string `mapstructure:"email"`
	APIKey   string `mapstructure:"apikey"`
	APIToken string `mapstructure:"apitoken"`
	Domain   string `mapstructure:"domain"`
}

type Cloudflare struct {
	config *Config
	client *cloudflare.API
	zoneID string
}

func init() {
	updater.Register("Cloudflare", func(v *viper.Viper) (updater.Updater, error) {
		cfg := Config{}
		err := v.UnmarshalKey("updaters.cloudflare", &cfg)
		if err != nil {
			return nil, err
		}
		return New(&cfg)
	})
}

func New(config *Config) (*Cloudflare, error) {
	var (
		api *cloudflare.API
		err error
	)

	if config.APIToken != "" {
		api, err = cloudflare.NewWithAPIToken(config.APIToken)
	} else {
		api, err = cloudflare.New(config.APIKey, config.Email)
	}

	if err != nil {
		slog.Debug("[CloudFlare] failed to create API client:", "error", err)
		return nil, err
	}

	return &Cloudflare{
		config: config,
		client: api,
	}, nil
}

func (c *Cloudflare) Update(ips *provider.IpResult) error {
	if c.zoneID == "" {
		domain := c.config.Domain
		parts := strings.Split(domain, ".")
		l := len(parts)
		zone := parts[l-2] + "." + parts[l-1]
		zoneID, err := c.client.ZoneIDByName(zone)
		if err != nil {
			return err
		}
		c.zoneID = zoneID
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	if ips.IPv4 != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.updateDNSRecord(ctx, "A", ips.IPv4); err != nil {
				errCh <- err
			}
		}()
	}

	if ips.IPv6 != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.updateDNSRecord(ctx, "AAAA", ips.IPv6); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cloudflare) updateDNSRecord(ctx context.Context, recordType, ip string) error {
	domain := c.config.Domain
	params := cloudflare.ListDNSRecordsParams{Type: recordType, Name: domain}
	dnsRecords, _, err := c.client.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(c.zoneID), params)
	if err != nil {
		slog.Error("[CloudFlare] failed to list DNS records:", "error", err, "type", recordType)
		c.zoneID = ""
		return err
	}

	if len(dnsRecords) > 0 {
		record := dnsRecords[0]
		updateParams := cloudflare.UpdateDNSRecordParams{
			ID:      record.ID,
			Type:    recordType,
			Name:    domain,
			Content: ip,
			TTL:     record.TTL,
			Proxied: record.Proxied,
		}

		_, err := c.client.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(c.zoneID), updateParams)
		if err != nil {
			slog.Error("[CloudFlare] failed to update DNS record:", "error", err, "type", recordType)
			return err
		}
		slog.Info("[CloudFlare] DNS record updated successfully", "type", recordType, "ip", ip)
	} else {
		createParams := cloudflare.CreateDNSRecordParams{
			Type:    recordType,
			Name:    domain,
			Content: ip,
			TTL:     60,
			Proxied: cloudflare.BoolPtr(false),
		}

		_, err := c.client.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(c.zoneID), createParams)
		if err != nil {
			slog.Error("[CloudFlare] failed to create DNS record:", "error", err, "type", recordType)
			return err
		}
		slog.Info("[CloudFlare] DNS record created successfully", "type", recordType, "ip", ip)
	}

	return nil
}
