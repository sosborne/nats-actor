package main

import (
	"flag"
	"fmt"

	"github.com/nats-io/nats"

	"github.com/sosborne/nats-actor"
)

func main() {
	var remote = flag.Bool("remote", false, "")
	flag.Parse()

	opts := nats.DefaultOptions
	conn, err := opts.Connect()
	if err != nil {
		panic(err)
	}

	recv := func(msg string, ac *actor.ActorContext) {
		if *remote {
			fmt.Printf("Remote received %s. Responding...\n.", msg)
			ac.Sender().Tell([]byte("Got your message!"), ctx.Self())
		} else {
			fmt.Println("Got response from remote", msg)
		}
	}

	as := actor.NewActorSystem("foosystem", conn)
	var name string
	if *remote {
		name = "remote"
	} else {
		name = "local"
	}

	
	a := as.ActorOf(name, recv)
	if !*remote {
		a.
	}
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
