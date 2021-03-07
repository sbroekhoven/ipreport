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

// GetOne functuin to get one ptr record
func GetOne(ip string, nameserver string) (string, error) {
	// var records []string

	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "", err
	}
	println("FIRST: " + strings.TrimRight(names[0], ".") + "\n")

	reversedIP, err := reverseIPv4(ip)
	println(reversedIP + "\n")

	var record string
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(reversedIP), dns.TypePTR)
	m.MsgHdr.RecursionDesired = true
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameserver+":53")
	if err != nil {
		return "", err
	}
	for _, rin := range in.Answer {
		if r, ok := rin.(*dns.PTR); ok {
			record = r.Ptr
			println("BLAA" + r.Ptr + " PTR\n")
		}
	}
	return record, nil
}

func reverseIPv4(ip string) (string, error) {
	PTR, err := dns.ReverseAddr(ip)
	if err != nil {
		return "", err
	}

	reversed := strings.TrimSuffix(PTR, ".in-addr.arpa.")

	return reversed, nil
}
