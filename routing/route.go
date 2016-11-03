package routing

import (
	"fmt"
	"net"
	"strings"
)

type Route interface {
	// could theoretically create dynamic targets for each match
	Match(network string, address AddressDetails) *Target
}

func ParseRoute(line string) (Route, error) {
	line = strings.TrimSpace(line)
	if 0 == len(line) || '#' == line[0] {
		return nil, nil
	}

	if '^' != line[0] {
		// drop trailing comment
		line = strings.Split(line, "#")[0]
		fields := strings.Fields(line)
		if 2 != len(fields) {
			return nil, fmt.Errorf("Invalid route: %q", line)
		}
		if target, err := ParseTarget(fields[1]); nil != err {
			return nil, err
		} else {
			return parseSimpleMatch(fields[0], target)
		}
	} else {
		return nil, fmt.Errorf("Regular expressions not supported yet: %q", line)
	}
}

type cidrRoute struct {
	CIDR   net.IPNet
	Port   string
	Target *Target
}

func (r cidrRoute) Match(network string, address AddressDetails) *Target {
	if nil != address.IP && r.CIDR.Contains(address.IP) && (0 == len(r.Port) || r.Port == address.Port) {
		return r.Target
	} else {
		return nil
	}
}

type domainRoute struct {
	Domain string
	Port   string
	Target *Target
}

func removeTrailingDot(fqdn string) string {
	// remove trailing dot
	if strings.HasSuffix(fqdn, ".") {
		return fqdn[:len(fqdn)-1]
	} else {
		return fqdn
	}
}

func (r domainRoute) Match(network string, address AddressDetails) *Target {
	fqdn := removeTrailingDot(address.FQDN)
	if 0 != len(fqdn) && (0 == len(r.Port) || r.Port == address.Port) {
		if "*" == r.Domain {
			return r.Target
		} else if '.' == r.Domain[0] {
			if r.Domain[1:] == fqdn || strings.HasSuffix(fqdn, r.Domain) {
				return r.Target
			}
		} else if r.Domain == fqdn {
			return r.Target
		}
	}
	return nil
}

func parseSimpleMatch(match string, target *Target) (Route, error) {
	var network string
	var host string
	var port string

	if '[' == match[0] {
		if rbracket := strings.LastIndex(match, "]"); -1 == rbracket {
			return nil, fmt.Errorf("Missing closing ']' in %q", match)
		} else {
			network = match[1:rbracket]
			if len(match) > rbracket+1 {
				if ':' != match[rbracket+1] {
					return nil, fmt.Errorf("Only ':' allowed after ']' in %q", match)
				}
				port = match[rbracket+1:]
			}
		}
	} else if slash := strings.IndexRune(match, '/'); -1 != slash {
		if lastcolon := strings.LastIndex(match, ":"); lastcolon > slash {
			network = match[:lastcolon]
			port = match[lastcolon+1:]
		} else {
			network = match
		}
	} else if colon := strings.IndexRune(match, ':'); -1 != colon {
		if lastcolon := strings.LastIndex(match, ":"); lastcolon > colon {
			// 2 or more colons: always interpret as IPv6 address
			network = match
		} else {
			// single colon: not an IPv6 address, so split port
			host = match[:colon]
			port = match[colon+1:]
		}
	} else {
		host = match
	}

	var ipnet net.IPNet
	if 0 != len(host) {
		// cannot be CIDR, but could be single IP address
		if ipnet.IP = net.ParseIP(host); nil != ipnet.IP {
			// single host, full mask
			ipnet.Mask = net.CIDRMask(8*len(ipnet.IP), 8*len(ipnet.IP))
			network = host
			host = ""
		}
	} else {
		if ipnet.IP = net.ParseIP(network); nil != ipnet.IP {
			// single host, full mask
			ipnet.Mask = net.CIDRMask(8*len(ipnet.IP), 8*len(ipnet.IP))
		} else if _, n, err := net.ParseCIDR(network); nil != err {
			return nil, fmt.Errorf("Invalid IP address/network: %q", network)
		} else {
			ipnet = *n
		}
	}

	if 0 != len(host) {
		return domainRoute{
			Domain: removeTrailingDot(host),
			Port:   port,
			Target: target,
		}, nil
	} else {
		return cidrRoute{
			CIDR:   ipnet,
			Port:   port,
			Target: target,
		}, nil
	}

	return nil, nil
}
