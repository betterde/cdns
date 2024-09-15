package challenge

import (
	"context"
	"github.com/betterde/cdns/pkg/dns"
	"github.com/mholt/acmez/v2/acme"
)

// Provider implements go-acme/lego Provider interface which is used for ACME DNS challenge handling
type Provider struct {
	servers []*dns.Server
}

// NewChallengeProvider creates a new instance of ChallengeProvider
func NewChallengeProvider(servers []*dns.Server) Provider {
	return Provider{servers: servers}
}

// Present is used for making the ACME DNS challenge token available for DNS
func (c *Provider) Present(ctx context.Context, chall acme.Challenge) error {
	for _, s := range c.servers {
		s.PersonalKeyAuth = chall.DNS01KeyAuthorization()
	}
	return nil
}

// CleanUp is called after the run to remove the ACME DNS challenge tokens from DNS records
func (c *Provider) CleanUp(ctx context.Context, challenge acme.Challenge) error {
	for _, s := range c.servers {
		s.PersonalKeyAuth = ""
	}
	return nil
}

// Wait is a dummy function as we are just going to be ready to answer the challenge from the get-go
func (c *Provider) Wait(ctx context.Context, challenge acme.Challenge) error {
	return nil
}
