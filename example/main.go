package main

import (
	"fmt"
	"time"

	"github.com/nats-io/nats"

	"github.com/sosborne/nats-actor"
)

func main() {
	opts := nats.DefaultOptions
	conn, err := opts.Connect()
	if err != nil {
		panic(err)
	}

	parentBecome := func(msg string, ac *actor.ActorContext) {
		fmt.Println("Parent become received", msg)
	}
	parentInitial := func(msg string, ac *actor.ActorContext) {
		fmt.Println("Parent initial received", msg)
		if msg == "become" {
			ac.Become(parentBecome)
		}
	}

	as := actor.NewActorSystem("foosystem", conn)
	a := as.ActorOf("parentActor", parentInitial)
	a.Self().Tell([]byte("Hello actor"), nil)
	a.Self().Tell([]byte("become"), nil)
	a.Self().Tell([]byte("Hello again actor"), nil)

	child := a.NewChild("childActor", func(msg string, _ *actor.ActorContext) {
		fmt.Println("Child received", msg)
	})
	child.Self().Tell([]byte("Hello child"), a.Self())

	conn.Flush()
	time.Sleep(1 * time.Second)
}
