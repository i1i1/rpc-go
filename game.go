package main

import (
	"bufio"
	"context"
	"fmt"
	"sync"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	multiaddr "github.com/multiformats/go-multiaddr"
)

type Host struct {
	host.Host
}

const PROTOCOL_ID = protocol.ID("/rpc-go/0.1")

var logger = log.Logger("rendezvous")

func (host *Host) handleStream(stream network.Stream) {
	logger.Info("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	//event_reader := go host.readData(rw)
	go host.writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func (host *Host) readData(rw *bufio.ReadWriter) chan<- Event {
	ch := make(chan Event)
	go func() {
		for {
			// TODO:
		}
	}()
	return ch
}

func (host *Host) writeData(rw *bufio.ReadWriter) chan<- Event {
	ch := make(chan Event)
	go func() {
		for ev := range ch {
			_ = ev
			fmt.Print("> ")
			// TODO:
		}
	}()
	return ch
}

func NewHost(ctx context.Context, listenAddresses addrList) (*Host, error) {
	p2phost, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]multiaddr.Multiaddr(listenAddresses)...),
	)
	if err != nil {
		return nil, err
	}
	host := &Host{Host: p2phost}
	host.Host = host
	host.SetStreamHandler(PROTOCOL_ID, func(s network.Stream) {
		host.handleStream(s)
	})
	return &Host{host}, nil
}

func (host *Host) Anounce(ctx context.Context, rendezvousString string) (*discovery.RoutingDiscovery, error) {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		return nil, err
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	logger.Debug("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				logger.Warning(err)
			} else {
				logger.Info("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	logger.Info("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, rendezvousString)

	return routingDiscovery, nil
}

func (host *Host) ConnectPeer(ctx context.Context, id peer.ID) error {
	stream, err := host.NewStream(ctx, id, PROTOCOL_ID)
	if err != nil {
		return err
	}

	host.handleStream(stream)
	return nil
}
