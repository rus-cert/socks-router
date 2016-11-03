package routing

import (
	"fmt"
	"net"
	"strings"
)

func ParseAddress(addr string) (*AddressDetails, error) {
	if host, port, err := net.SplitHostPort(addr); nil != err {
		return nil, err
	} else {
		zoneParts := strings.Split(host, "%")
		var zone string
		if 2 == len(zoneParts) {
			host = zoneParts[0]
			zone = zoneParts[1]
		} else if 1 != len(zoneParts) {
			return nil, fmt.Errorf("Invalid Host %q", host)
		}
		ip := net.ParseIP(host)
		if ip != nil {
			host = ""
		}
		return &AddressDetails{
			Address: addr,
			FQDN:    host,
			IP:      ip,
			Zone:    zone,
			Port:    port,
		}, nil
	}
}

type AddressDetails struct {
	Address string // for dial functions: [...]:...
	FQDN    string // only the domain name, if available
	IP      net.IP
	Zone    string // ipv6 zone [...%zone]:...
	Port    string
}
