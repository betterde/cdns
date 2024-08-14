package handler

import (
	"github.com/betterde/cdns/internal/response"
	"github.com/betterde/cdns/pkg/dns"
	"github.com/gofiber/fiber/v2"
	record "github.com/miekg/dns"
	"strings"
)

type Request struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

// Present create TXT record
func Present(ctx *fiber.Ctx) error {
	payload := Request{}
	err := ctx.BodyParser(&payload)
	if err != nil {
		return ctx.JSON(response.ValidationError("Payload validation failed.", err))
	}

	for _, server := range dns.Servers {
		txtRecord := &record.TXT{
			Hdr: record.RR_Header{
				Name:   payload.FQDN,
				Rrtype: record.TypeTXT,
				Class:  record.ClassINET,
				Ttl:    3600,
			},
			Txt: []string{payload.Value},
		}

		domain := server.Domains[payload.FQDN]
		domain.Records = append(server.Domains[payload.FQDN].Records, txtRecord)
		server.Domains[payload.FQDN] = domain
	}

	return ctx.JSON(response.Success("Success", nil))
}

// Cleanup delete TXT record
func Cleanup(ctx *fiber.Ctx) error {
	payload := Request{}
	err := ctx.BodyParser(&payload)
	if err != nil {
		return ctx.JSON(response.ValidationError("Payload validation failed.", err))
	}

	for _, server := range dns.Servers {
		domain := server.Domains[payload.FQDN]
		result := make([]record.RR, 0)
		for _, rec := range domain.Records {
			txtSlice := rec.(*record.TXT).Txt
			if strings.Join(txtSlice, "") != payload.Value {
				result = append(result, rec)
			}
		}
		server.Domains[payload.FQDN] = dns.Records{Records: result}
	}

	return ctx.JSON(response.Success("Success", nil))
}
