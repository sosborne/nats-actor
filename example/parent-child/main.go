package main

import (
	"fmt"
	"sync"

	"github.com/nats-io/nats"

	"github.com/sosborne/nats-actor"
)

func main() {
	opts := nats.DefaultOptions
	conn, err := opts.Connect()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	parentBecome := func(msg string, ac *actor.ActorContext) {
		fmt.Println("Parent become received", msg)
		wg.Done()
	}
	parentInitial := func(msg string, ac *actor.ActorContext) {
		fmt.Println("Parent initial received", msg)
		if msg == "become" {
			ac.Become(parentBecome)
		}
	}

	as := actor.NewActorSystem("foosystem", conn)
	a := as.ActorOf("parentActor", parentInitial)
	wg.Add(1)
	a.Self().Tell([]byte("Hello actor"), nil)
	a.Self().Tell([]byte("become"), nil)
	a.Self().Tell([]byte("Hello again actor"), nil)

	child := a.NewChild("childActor", func(msg string, _ *actor.ActorContext) {
		fmt.Println("Child received", msg)
		wg.Done()
	})
	wg.Add(1)
	child.Self().Tell([]byte("Hello child"), a.Self())

	conn.Flush()
	wg.Wait()
}
