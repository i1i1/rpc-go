package main

import (
	"context"
	"flag"
)

func main() {
	initLogger()

	help := flag.Bool("h", false, "Display Help")
	config, err := ParseFlags()
	if err != nil {
		panic(err)
	}

	if *help {
		printHelp()
		return
	}

	ctx := context.Background()

	host, err := NewHost(ctx, config.ListenAddresses)
	if err != nil {
		panic(err)
	}
	logger.Info("Host created. We are:", host.ID())
	logger.Info(host.Addrs())

	routingDiscovery, err := host.Anounce(ctx, config.RendezvousString)
	if err != nil {
		panic(err)
	}
	logger.Debug("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	logger.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
	if err != nil {
		panic(err)
	}

	for peer := range peerChan {
		if peer.ID == host.ID() {
			continue
		}
		logger.Debug("Found peer:", peer)
		logger.Debug("Connecting to:", peer)
		err := host.ConnectPeer(ctx, peer.ID)

		if err != nil {
			logger.Warning("Connection failed:", err)
		} else {
			logger.Info("Connected to:", peer)
		}
	}

	select {}
}
