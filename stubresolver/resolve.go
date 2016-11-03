package stubresolver

import (
	"golang.org/x/net/context"
	"net"
)

/* stores the name in the context, doesn't actually resolve;
 * extract with FqdnFromContext
 */
type StubResolver struct{}

func (StubResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	return NewFqdnContext(ctx, name), nil, nil
}
