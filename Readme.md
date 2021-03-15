# Distributed Systems and Middleware Final Project

## Innopolis University, Spring 2021

## 1. Description

This is a distributed game that uses peer-to-peer architecture. Each player is
a peer.  
The game uses Rock-Paper-Scissors-Lizard-Spock rules, originally described
[there][rules-origin].

## 2. How does it work

We use libp2p. What is Libp2p?  
[Good question!][what-libp2p] Libp2p is a set of protocols, and libraries, that
allows to easily build peer-to-peer applications. In other words, it is a
**middleware** for peer-to-peer systems.  

libp2p peers are addressed with combination of unique PeerId and listening
address, which is regular TCP address/port combination.

libp2p manages communication between peers. It provides a convenient
publish-subscribe channel abstraction. We use gossip protocol to make messages
reach all peers.

Games hold between peers in the same network. Peers can discover each other
with mDNS. When a host starts, it creates a PubSub channel, that will be used
for game. It will automatically receives new messages from the game topic.

## 2. Game process

When a peer wants to launch a new game, it publishes a special message into the
PubSub topic. To start a game, all players who wants to participate should send
a vote for starting game.

If majority of players will not respond in a certain amount of time, game will
not start. If all nodes voted for starting a game, game starts. Players who 
had not voted can be kicked.

When the game starts, each node should do following:

* create an ECDSA key
* choose his move (Rock, Paper, etc.)
* send message with encrypted move

After node receives moves from all other players, it sends the key to PubSub.
If some node did not send move or key within a certain timeout, it can be
kicked.

Kicking forbids player to take part in games in this GameRoom. Node can start
voting for kicking other node by sending a special message to PubSub. Then all
of the players can vote within a certain timeout. If majority votes. node will
be ignored and no players will accept messages from it.

When node knows moves of all players, it performs topological sort to determine
who won a game.

<!-- References -->
[rules-origin]:http://www.samkass.com/theories/RPSSL.html
[what-libp2p]:https://docs.libp2p.io/introduction/what-is-libp2p/