package main

import (
	"log"
	"os"

	"github.com/samuelventura/go-state"
	"github.com/samuelventura/go-tools"
	"github.com/samuelventura/go-tree"
)

func main() {
	tools.SetupLog()

	ctrlc := tools.SetupCtrlc()
	stdin := tools.SetupStdinAll()

	log.Println("start", os.Getpid())
	defer log.Println("exit")

	rnode := tree.NewRoot("root", log.Println)
	defer rnode.WaitDisposed()
	//recover closes as well
	defer rnode.Recover()
	rnode.SetValue("name", tools.GetEnviron("SHIP_NAME", tools.GetHostname()))
	rnode.SetValue("keypath", tools.GetEnviron("SHIP_DOCK_KEYPATH", tools.WithExtension("key")))
	rnode.SetValue("pool", tools.GetEnviron("SHIP_DOCK_POOL", "127.0.0.1:31622"))
	rnode.SetValue("record", tools.GetEnviron("SHIP_DOCK_RECORD", ""))
	rnode.SetValue("state", tools.GetEnviron("SHIP_STATE", tools.WithExtension("state")))

	snode := state.Serve(rnode, rnode.GetValue("state").(string))
	defer snode.WaitDisposed()
	defer snode.Close()

	anode := rnode.AddChild("api")
	defer anode.WaitDisposed()
	defer anode.Close()
	run(anode)

	select {
	case <-rnode.Closed():
	case <-snode.Closed():
	case <-anode.Closed():
	case <-ctrlc:
	case <-stdin:
	}
}
