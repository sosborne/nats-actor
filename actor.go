package actor

import (
	"sync"

	"github.com/nats-io/nats"
)

// Receive is the operation an Actor performs upon the
// receipt of a message.
type Receive func(msg string, ctx *ActorContext)

// ActorContext provides contexual information for the actor.
// It also exposes the ability to Become - i.e. change
// behavior on the next received message.
type ActorContext struct {
	intial    Receive
	current   Receive
	sender    *ActorRef
	receiveMu sync.RWMutex
	senderMu  sync.Mutex
}

// Become changes the actor's behavior on the next received message.
func (ac *ActorContext) Become(receive Receive) {
	ac.receiveMu.Lock()
	ac.current = receive
	ac.receiveMu.Unlock()
}

// Gets the actor's ActorRef
func (ac *ActorContext) Self() *ActorRef {
	return ac.self
}

// Get the ActerRef that sent the message.
func (ac *ActorContext) Sender() *ActorRef {
	ac.senderMu.Lock()
	s := ac.sender
	ac.senderMu.Unlock()
	return s
}

func (ac *ActorContext) getCurrentBehavior() Receive {
	ac.receiveMu.RLock()
	r := ac.current
	ac.receiveMu.RUnlock()
	return r
}

func (ac *ActorContext) setSender(sender *ActorRef) {
	ac.senderMu.Lock()
	ac.sender = sender
	ac.senderMu.Unlock()
}

// Immutible and serializable handle to an Actor.
type ActorRef struct {
	mailbox string
	system  *ActorSystem
}

// Tell sends the message to the given actor.
func (ar *ActorRef) Tell(msg []byte, sender *ActorRef) {
	if sender != nil {
		ar.system.conn.PublishRequest(ar.mailbox, sender.mailbox, msg)
	} else {
		ar.system.conn.Publish(ar.mailbox, msg)
	}
}

// Actor is the basic unit of computation in the ActorSytem.
type Actor struct {
	Name       string
	parent     *Actor
	ctx        *ActorContext
	children   []*Actor
	system     *ActorSystem
	self       *ActorRef
	mailboxSub *nats.Subscription
	mailboxC   chan *nats.Msg
	killC      chan struct{}
}

// NewChild creates a new child actor for the actor.
func (a *Actor) NewChild(name string, receive Receive) *Actor {
	child := a.system.actorWithParent(name, receive, a)
	a.children = append(a.children, child)
	return child
}

// Self returns the ActorRef for the actor.
func (a *Actor) Self() *ActorRef {
	return a.self
}

func (a *Actor) handler(msg *nats.Msg) {
	a.mailboxC <- msg
}

func (a *Actor) start() {
	// TODO: Do something with this error
	a.mailboxSub, _ = a.system.conn.Subscribe(a.self.mailbox, a.handler)
	go func() {
		for {
			select {
			case msg := <-a.mailboxC:
				if msg.Reply != "" {
					a.ctx.setSender(&ActorRef{
						mailbox: msg.Reply,
						system:  a.system,
					})
				}
				a.ctx.getCurrentBehavior()(string(msg.Data), a.ctx)
			case <-a.killC:
				return
			}
		}
	}()
}

func (a *Actor) stop() {
	if a.mailboxSub != nil {
		// TODO: Do something with this error
		a.mailboxSub.Unsubscribe()
	}
	a.killC <- struct{}{}
}
