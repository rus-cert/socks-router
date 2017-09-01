# socks-router

`socks-router` is a proxy router; routing is configured based on
hostnames and IP addresses.

It accepts requests using any of the following protocols:
- SOCKS4, SOCKS4a, SOCKS5
- HTTP (CONNECT and normal request methods)
- CONNECT (similar to HTTP CONNECT, used e.g. by openssl)

It can forward requests to a SOCKS5 proxy or use a direct TCP
connection.

## Build

Build using go:

    go get github.com/rus-cert/socks-router

There is also a systemd service file which assumes the binary ends up in
`/usr/bin` and the config in `/etc/socks-router.routes` and listens to
`127.0.0.1:1080` and `[::1]:1080`.  After modifying (or creating) the
file you need to restart the service.

## Usecase

A typical usecase would be establishing a SSH-connection with a
"DynamicForward 127.0.0.1:PORT" option.  Now you can connect to systems
through the SOCKS5-proxy listening on 127.0.0.1:PORT as if you'd be on
the SSH target machine.  That way you can access internal systems behind
firewalls.

Now you don't want to route all traffic through the proxy (e.g. by
setting the proxy property in your web browser), as you might want to
use the application normally too if you are not connected, or have
multiple SSH-connections to different sites.

Now you can setup `socks-router` to run a local proxy and route only
requests for "protected" hostnames / IP addresses through the SSH proxy,
and use direct connections for everything else.

### Example

`~/.ssh/config`

    Host gateway.example.com
        DynamicForward 127.0.0.1:2080

`/etc/socks-router.routes` or `~/socks-routes`

    10.40.40.0/24     socks5://127.0.0.1:2080
    .example.com      socks5://127.0.0.1:2080

## Configuration

The configuration consists of a list of routing rules.

Each line is handled as follow:

- leading and trailing space is ignored
- '#' till end of line marks a comment
- empty lines are ignored
- '^' at the beginning marks regular expression match; not supported yet
  (might handle '#' differently)
- otherwise each line contains a "match" and a "target" column separated by whitespace
- valid matches:
  - [ip-addr/prefix]:port
    "/prefix" and ":port" are optional
  - ip-addr/prefix:port
    ":port" is optional
  - ipv6-addr
    Needs to contain at least two colons. Cannot specify a port that way!
  - ipv4-addr:port
  - domain.name:port
    ":port" is optional
    trailing dots are ignored
    leading dot means match domain and all subdomains
  - *:port
    ":port" is optional
    matches all addresses and domain names
- valid targets:
  - socks5://address:port
  - direct

## Routing

Each request either uses a hostname or an IP address; `socks-router`
does not try to resolve hostnames into IP addresses or vice versa for
routing.  This means that a request using a hostname is only routed
through "hostname rules" (and the wildcard '*'), and a request using an
IP address only routed using IP-address based rules.

The IP address matching uses `net.IPNet.Contains`, which interprets IPv4
as part of IPv6 by padding it with zeroes on the left.

If no rule matched the default is to route "direct", i.e. using a local
TCP connection.

## Application configuration

Some applications have dedicated configurations for proxy settings
(Firefox for example), and might default to "system default" or "direct
connection".  Other applications (Chrome) always use the "system
default".

"system default" is a tricky thing though; there doesn't seem to exist a
single standard how to specify it, although it mostly is based on the
"HTTP_PROXY" environment variable.

Depending on the application it might expect a protocol in the value
like this:

    HTTP_PROXY="http://127.0.0.1:1080"

or, if they support SOCKS5 as well:

    HTTP_PROXY="socks5://127.0.0.1:1080"

Others might only want an address and port like:

    HTTP_PROXY="127.0.0.1:1080"

The "icedtea" implementation for "Java Web Start" (JNLP) can be
configured through the `itweb-settings` application.
