package ptr

import (
	"net"
	"strings"

	"github.com/miekg/dns"
)

/*
 * Get ptr
 * TODO: Rewrite
 */

// Get function that get the nameservers of a domain
func Get(domain string, nameserver string) ([]string, error) {
	var answer []string
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypePTR)
	m.MsgHdr.RecursionDesired = true
	// m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameserver+":53")
	if err != nil {
		return answer, err
	}
	for _, rin := range in.Answer {
		if r, ok := rin.(*dns.PTR); ok {
			answer = append(answer, r.Ptr)
		}
	}
	if len(answer) < 1 {
		return answer, err
	}
	return answer, nil
}

// GetOne functuin to get one ptr record
func GetOne(ip string, nameserver string) (string, error) {
	// var records []string

	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "", err
	}
	println(strings.TrimRight(names[0], "."))

	var record string
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(ip), dns.TypePTR)
	m.MsgHdr.RecursionDesired = true
	// m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameserver+":53")
	if err != nil {
		return "", err
	}
	for _, rin := range in.Answer {
		if r, ok := rin.(*dns.PTR); ok {
			record = r.Ptr
			println(r.Ptr)
		}
	}
	return record, nil
}
