package ptr

import (
	"github.com/miekg/dns"
)

/*
 * Get ptr
 * TODO: Rewrite
 */

// GetOne functuin to get one ptr record
func GetOne(ip string, nameserver string) (string, error) {
	reversedIP, err := dns.ReverseAddr(ip)
	if err != nil {
		return "", err
	}

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
		}
	}
	return record, nil
}
