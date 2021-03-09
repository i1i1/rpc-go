package game

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/i1i1/rpc-go/pkg/events"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// GameRoomBufSize is the number of incoming messages to buffer for each topic.
const GameRoomBufSize = 128

type (
	// GameRoom represents a subscription to a single PubSub topic. Events
	// can be published to the topic with GameRoom.Publish, and received
	// events are pushed to the Events channel.
	GameRoom struct {
		// Events is a channel of events received from other peers in the game room
		Events chan events.Event
		// Ctx is a game context
		Ctx context.Context
		// RoomName is name of the room
		RoomName RoomName
		// Self is our player name and id
		Self events.Player

		ps    *pubsub.PubSub
		topic *pubsub.Topic
		sub   *pubsub.Subscription

		bannedUsers map[peer.ID]struct{}
	}

	RoomName string

	SendEvent struct {
		Type  events.EventType
		Event events.Event
	}
)

func (name *RoomName) topicName() string {
	return "chat-room:" + string(*name)
}

// JoinGameRoom tries to subscribe to the PubSub topic for the room name, returning
// a GameRoom on success.
func JoinGameRoom(
	ctx context.Context,
	ps *pubsub.PubSub,
	selfID peer.ID,
	nickname string,
	roomName RoomName,
) (*GameRoom, error) {
	// join the pubsub topic
	topic, err := ps.Join(roomName.topicName())
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	gr := &GameRoom{
		Ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		Self:     events.Player{ID: selfID, Nick: nickname},
		RoomName: roomName,
		Events:   make(chan events.Event, GameRoomBufSize),
	}

	// start reading messages from the subscription in a loop
	go gr.readLoop()
	return gr, nil
}

// Publish sends a message to the pubsub topic.
func (gr *GameRoom) Publish(event events.Event) error {
	ev := SendEvent{
		Event: event,
		Type:  event.Type(),
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&ev); err != nil {
		return err
	}
	return gr.topic.Publish(gr.Ctx, buf.Bytes())
}

func (gr *GameRoom) ListPeers() map[peer.ID]struct{} {
	out := map[peer.ID]struct{}{}
	for _, p := range gr.ps.ListPeers(gr.RoomName.topicName()) {
		if _, ok := gr.bannedUsers[p]; !ok {
			out[p] = struct{}{}
		}
	}
	return out
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (gr *GameRoom) readLoop() {
	for {
		msg, err := gr.sub.Next(gr.Ctx)
		if err != nil {
			close(gr.Events)
			return
		}
		// only forward events delivered by others
		if msg.ReceivedFrom == gr.Self.ID {
			continue
		}

		event := SendEvent{}
		reader := bytes.NewReader(msg.Data)
		dec := gob.NewDecoder(reader)
		err = dec.Decode(&event)
		if err != nil {
			fmt.Printf("Err: %v\n", err)
			continue
		}

		if event.Event.Type() != event.Type {
			// TODO: Print errors
			continue
		}

		gr.Events <- event.Event
	}
}
