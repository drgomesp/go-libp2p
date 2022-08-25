package swarm

import (
	"context"
	"crypto/rand"
	"net"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/test"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	"github.com/stretchr/testify/require"
)

func TestAddrsForDial(t *testing.T) {
	mockResolver := madns.MockResolver{IP: make(map[string][]net.IPAddr)}
	ipaddr, err := net.ResolveIPAddr("ip4", "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	mockResolver.IP["example.com"] = []net.IPAddr{*ipaddr}

	resolver, err := madns.NewResolver(madns.WithDomainResolver("example.com", &mockResolver))
	if err != nil {
		t.Fatal(err)
	}

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)
	id, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	ps, err := pstoremem.NewPeerstore()
	require.NoError(t, err)
	ps.AddPubKey(id, priv.GetPublic())
	ps.AddPrivKey(id, priv)
	t.Cleanup(func() { ps.Close() })

	tpt, err := websocket.New(nil, network.NullResourceManager)
	require.NoError(t, err)
	s, err := NewSwarm(id, ps, WithMultiaddrResolver(resolver))
	require.NoError(t, err)
	err = s.AddTransport(tpt)
	require.NoError(t, err)

	otherPeer := test.RandPeerIDFatal(t)

	ps.AddAddr(otherPeer, multiaddr.StringCast("/dns4/example.com/tcp/1234/wss"), time.Hour)

	ctx := context.Background()
	mas, err := s.addrsForDial(ctx, otherPeer)
	require.NoError(t, err)

	require.NotZero(t, len(mas))
}
