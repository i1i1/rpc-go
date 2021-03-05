package main

type (
	EventType int

	Event interface {
		Submit()
		Respond()
	}

	EventStartGameVote struct{}
	EventStartGame     struct{}
)

const (
	// Event which starts voting for game
	EVENT_START_GAME_VOTE EventType = iota
	// Event which happends during vote for game.
	// It means that peer is agreed to start game
	EVENT_START_GAME
	// Event which starts voting for kick of some peer
	EVENT_START_KICK_VOTE
	// Event which happends during vote of the peer.
	// Event which coresponds to agreement of peer to kick the another one
	EVENT_KICK_VOTE
)
