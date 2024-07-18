package dns

import (
	"fmt"
	"github.com/betterde/cdns/config"
	"github.com/betterde/cdns/internal/journal"
	"github.com/miekg/dns"
	"net"
	"strings"
	"time"
)

var Servers []*Server

// Records is a slice of ResourceRecords
type Records struct {
	Records []dns.RR
}

// Server is the main struct for acme-dns DNS server
type Server struct {
	A               dns.RR
	SOA             dns.RR
	Domain          string
	Server          *dns.Server
	Domains         map[string]Records
	PersonalKeyAuth string
}

func InitServer(errChan chan error) {
	servers := make([]*Server, 0)

	if strings.HasPrefix(config.Conf.DNS.Protocol, "both") {
		// Handle the case where DNS server should be started for both udp and tcp
		udpProto := "udp"
		tcpProto := "tcp"
		if strings.HasSuffix(config.Conf.DNS.Protocol, "4") {
			udpProto += "4"
			tcpProto += "4"
		} else if strings.HasSuffix(config.Conf.DNS.Protocol, "6") {
			udpProto += "6"
			tcpProto += "6"
		}

		udpServer := newServer(config.Conf.DNS.Listen, udpProto)
		servers = append(servers, udpServer)

		tcpServer := newServer(config.Conf.DNS.Listen, tcpProto)
		servers = append(servers, tcpServer)

		// No need to parse records from config again
		tcpServer.SOA = udpServer.SOA

		go udpServer.Start(errChan)
		go tcpServer.Start(errChan)
	} else {
		dnsServer := newServer(config.Conf.DNS.Listen, config.Conf.DNS.Protocol)
		servers = append(servers, dnsServer)
		go dnsServer.Start(errChan)
	}

	Servers = servers
}

func newServer(addr, proto string) *Server {
	var server Server
	server.Server = &dns.Server{Addr: addr, Net: proto}

	domain := config.Conf.HTTP.Domain
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	server.Domain = strings.ToLower(domain)
	server.Domains = make(map[string]Records)

	serial := time.Now().Format("2006010215")
	// Add SOA
	SOAStr := fmt.Sprintf("%s. SOA %s. %s. %s 28800 7200 604800 86400", strings.ToLower(config.Conf.SOA.Domain), strings.ToLower(config.Conf.DNS.NSName), strings.ToLower(config.Conf.DNS.Admin), serial)
	SOARR, err := dns.NewRR(SOAStr)
	if err != nil {
		journal.Logger.With("Error", err.Error(), "SOA", SOAStr).Error("Error while adding SOA record")
	} else {
		server.appendRR(SOARR)
		server.SOA = SOARR
	}

	return &server
}

// Start starts the DNSServer
func (d *Server) Start(errorChannel chan error) {
	// DNS server part
	dns.HandleFunc(".", d.handleRequest)

	journal.Logger.With("Addr", d.Server.Addr, "Proto", d.Server.Net).Debug("Listening DNS")

	err := d.Server.ListenAndServe()
	if err != nil {
		errorChannel <- err
	}
}

func (d *Server) appendRR(rr dns.RR) {
	addDomain := rr.Header().Name
	_, ok := d.Domains[addDomain]
	if !ok {
		d.Domains[addDomain] = Records{[]dns.RR{rr}}
	} else {
		domain := d.Domains[addDomain]
		domain.Records = append(domain.Records, rr)
		d.Domains[addDomain] = domain
	}
	journal.Logger.With("Domain", addDomain, "RecordType", dns.TypeToString[rr.Header().Rrtype]).Debug("Adding new record to domain")
}

func (d *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	// handle DNS0
	opt := r.IsEdns0()
	if opt != nil {
		if opt.Version() != 0 {
			// Only DNS0 is standardized
			m.MsgHdr.Rcode = dns.RcodeBadVers
			m.SetEdns0(512, false)
		} else {
			// We can safely do this as we know that we're not setting other OPT RRs within acme-dns.
			m.SetEdns0(512, false)
			if r.Opcode == dns.OpcodeQuery {
				d.readQuery(m)
			}
		}
	} else {
		if r.Opcode == dns.OpcodeQuery {
			d.readQuery(m)
		}
	}
	_ = w.WriteMsg(m)
}

func (d *Server) readQuery(m *dns.Msg) {
	var authoritative = false
	for _, que := range m.Question {
		if rr, rc, auth, err := d.answer(que); err == nil {
			if auth {
				authoritative = auth
			}
			m.MsgHdr.Rcode = rc
			m.Answer = append(m.Answer, rr...)
		}
	}
	m.MsgHdr.Authoritative = authoritative
	if authoritative {
		if m.MsgHdr.Rcode == dns.RcodeNameError {
			m.Ns = append(m.Ns, d.SOA)
		}
	}
}

func (d *Server) getRecord(q dns.Question) ([]dns.RR, error) {
	var rr []dns.RR
	var cnames []dns.RR
	domain, ok := d.Domains[strings.ToLower(q.Name)]
	if !ok {
		return rr, fmt.Errorf("No records for domain %s", q.Name)
	}
	for _, ri := range domain.Records {
		if ri.Header().Rrtype == q.Qtype {
			rr = append(rr, ri)
		}
		if ri.Header().Rrtype == dns.TypeCNAME {
			cnames = append(cnames, ri)
		}
	}
	if len(rr) == 0 {
		return cnames, nil
	}
	return rr, nil
}

// answeringForDomain checks if we have any records for a domain
func (d *Server) answeringForDomain(name string) bool {
	if d.Domain == strings.ToLower(name) {
		return true
	}
	_, ok := d.Domains[strings.ToLower(name)]
	return ok
}

func (d *Server) isAuthoritative(q dns.Question) bool {
	if d.answeringForDomain(q.Name) {
		return true
	}
	domainParts := strings.Split(strings.ToLower(q.Name), ".")
	for i := range domainParts {
		if d.answeringForDomain(strings.Join(domainParts[i:], ".")) {
			return true
		}
	}
	return false
}

// isOwnChallenge checks if the query is for the domain of this acme-dns instance. Used for answering its own ACME challenges
func (d *Server) isOwnChallenge(name string) bool {
	domainParts := strings.SplitN(name, ".", 2)
	if len(domainParts) == 2 {
		if strings.ToLower(domainParts[0]) == "_acme-challenge" {
			domain := strings.ToLower(domainParts[1])
			if !strings.HasSuffix(domain, ".") {
				domain = domain + "."
			}
			if domain == d.Domain {
				return true
			}
		}
	}
	return false
}

func (d *Server) answer(q dns.Question) ([]dns.RR, int, bool, error) {
	var rcode int
	var err error
	var txtRRs []dns.RR
	var authoritative = d.isAuthoritative(q)
	if !d.isOwnChallenge(q.Name) && !d.answeringForDomain(q.Name) {
		rcode = dns.RcodeNameError
	}
	r, _ := d.getRecord(q)

	if q.Qtype == dns.TypeA && len(r) == 0 {
		var ip net.IP
		if q.Name == fmt.Sprintf("%s.", config.Conf.DNS.NSName) {
			ip = net.ParseIP(config.Conf.NS.IP)
		} else {
			ip = net.ParseIP(config.Conf.Ingress.IP)
		}

		r = append(r, &dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			A: ip,
		})
	}

	if q.Qtype == dns.TypeNS && len(r) == 0 {
		r = append(r, &dns.NS{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeNS,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Ns: fmt.Sprintf("%s.", config.Conf.DNS.NSName),
		})
	}

	if q.Qtype == dns.TypeTXT {
		if d.isOwnChallenge(q.Name) {
			txtRRs, err = d.answerOwnChallenge(q)
		}

		if err == nil {
			r = append(r, txtRRs...)
		}
	}
	if len(r) > 0 {
		// Make sure that we return NOERROR if there were dynamic records for the domain
		rcode = dns.RcodeSuccess
	}

	journal.Logger.With("QType", dns.TypeToString[q.Qtype], "Domain", q.Name, "RCode", dns.RcodeToString[rcode]).Debug("Answering question for domain")

	return r, rcode, authoritative, nil
}

// answerOwnChallenge answers to ACME challenge for acme-dns own certificate
func (d *Server) answerOwnChallenge(q dns.Question) ([]dns.RR, error) {
	r := new(dns.TXT)
	r.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 1}
	r.Txt = append(r.Txt, d.PersonalKeyAuth)
	return []dns.RR{r}, nil
}
