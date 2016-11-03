package stubresolver

import (
	"golang.org/x/net/context"
)

// key is an unexported type for keys defined in this package. This
// prevents collisions with keys defined in other packages.
type fqdnKey int

// userFqdnKey is the key for fqdn values in Contexts.  It is
// unexported; clients use NewFqdnContext and FqdnFromContext instead of
// using this key directly.
const userFqdnKey fqdnKey = 0

// NewFqdnContext returns a new Context that carries fqdn.
func NewFqdnContext(ctx context.Context, fqdn string) context.Context {
	return context.WithValue(ctx, userFqdnKey, &fqdn)
}

// FqdnFromContext returns the Fqdn value stored in ctx, if any.
func FqdnFromContext(ctx context.Context) (string, bool) {
	if val := ctx.Value(userFqdnKey); nil == val {
		return "", false
	} else if fqdn, ok := val.(*string); ok && nil != fqdn {
		return *fqdn, true
	} else {
		return "", false
	}
}
