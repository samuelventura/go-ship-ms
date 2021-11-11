package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/samuelventura/go-state"
	"github.com/samuelventura/go-tree"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)

	log.Println("start", os.Getpid())
	defer log.Println("exit")

	rlog := tree.NewLog()
	rnode := tree.NewRoot("root", rlog)
	defer rnode.WaitDisposed()
	//recover closes as well
	defer rnode.Recover()

	spath := state.SingletonPath()
	snode := state.Serve(rnode, spath)
	defer snode.WaitDisposed()
	defer snode.Close()
	log.Println("socket", spath)

	anode := rnode.AddChild("api")
	defer anode.WaitDisposed()
	defer anode.Close()
	anode.SetValue("name", getenv("SHIP_NAME", hostname()))
	anode.SetValue("keypath", getenv("SHIP_DOCK_KEYPATH", withext("key")))
	anode.SetValue("pool", getenv("SHIP_DOCK_POOL", "127.0.0.1:31622"))
	anode.SetValue("record", getenv("SHIP_DOCK_RECORD", ""))
	run(anode)

	stdin := make(chan interface{})
	go func() {
		defer close(stdin)
		ioutil.ReadAll(os.Stdin)
	}()
	select {
	case <-rnode.Closed():
	case <-snode.Closed():
	case <-anode.Closed():
	case <-ctrlc:
	case <-stdin:
	}
}
