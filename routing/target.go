package routing

import (
	"fmt"
	"net"

	"golang.org/x/net/proxy"
)

type Target struct {
	Name   string
	Dialer Dialer
}

var DirectTarget = Target{
	Name: "direct",
	Dialer: &net.Dialer{
		// enable RFC 6555-compliant "Happy Eyeballs"
		DualStack: true,
	},
}

func ParseTarget(name string) (*Target, error) {
	if "direct" == name {
		return &DirectTarget, nil
	} else if "socks5://" == name[0:9] {
		if dial, err := proxy.SOCKS5("tcp", name[9:], nil /* auth */, proxy.Direct); nil != err {
			return nil, fmt.Errorf("Invalid target: %v", err)
		} else {
			return &Target{
				Name:   name,
				Dialer: dial,
			}, nil
		}
	} else {
		return nil, fmt.Errorf("Invalid target: %q", name)
	}
}
