package actor

import (
	"fmt"

	"github.com/nats-io/nats"
)

// NewActorSystem creates a new ActorSystem with the given name
// and Nats Connection.
func NewActorSystem(name string, conn *nats.Conn) *ActorSystem {
	return &ActorSystem{
		name:         name,
		conn:         conn,
		systemActors: []*Actor{},
	}
}

// ActorSystem is a hierarchical group of actor which share common
// configuration. It is also the entry point for creating actors.
type ActorSystem struct {
	name         string
	conn         *nats.Conn
	systemActors []*Actor
}

// ActorOf creates a new Actor with the given name and receive handler.
func (as *ActorSystem) ActorOf(name string, receive Receive) *Actor {
	a := as.actorWithParent(name, receive, nil)
	as.systemActors = append(as.systemActors, a)
	return a
}

func (as *ActorSystem) actorWithParent(name string, receive Receive, parent *Actor) *Actor {
	a := &Actor{
		Name:   name,
		parent: parent,
		ctx: &ActorContext{
			intial:  receive,
			current: receive,
		},
		children: []*Actor{},
		system:   as,
		self: &ActorRef{
			mailbox: fmt.Sprintf("%s:%s", as.name, name),
			system:  as,
		},
		mailboxC: make(chan *nats.Msg),
		killC:    make(chan struct{}),
	}
	a.start()
	return a
}
