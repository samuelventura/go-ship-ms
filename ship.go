package main

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/samuelventura/go-tools"
	"github.com/samuelventura/go-tree"
	"golang.org/x/crypto/ssh"
)

func run(node tree.Node) {
	name := node.GetValue("name").(string)
	pool := node.GetValue("pool").(string)
	record := node.GetValue("record").(string)
	count := node.GetValue("count").(*countDso)
	keypath := node.GetValue("keypath").(string)
	id := node.GetValue("id").(*idDso)
	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		log.Panicln(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Panicln(err)
	}
	hkcb := func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }
	config := &ssh.ClientConfig{
		User:            name,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.HostKeyCallback(hkcb),
	}
	var txts []string
	if len(record) > 0 {
		txts, err = net.LookupTXT(record)
		if err != nil {
			log.Panicln(err)
		}
	} else {
		txts = []string{pool}
	}
	var eps []string
	for _, txt := range txts {
		addrs := strings.Split(txt, ",")
		eps = append(eps, addrs...)
	}
	log.Println("eps", eps)
	rand.Seed(time.Now().UnixNano())
	l := len(eps)
	//random start
	n := rand.Intn(l)
	for i := 0; i < l; i++ {
		ii := (n + i) % l
		addr := eps[ii]
		log.Println("ep", ii, addr, name)
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			log.Println(err)
			continue
		}
		tools.KeepAlive(conn, 5)
		sshCon, sshch, reqch, err := ssh.NewClientConn(conn, addr, config)
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		node.AddCloser("conn", conn.Close)
		node.AddCloser("sshCon", sshCon.Close)
		node.AddProcess("ping", func() {
			defer log.Println("request handler exited")
			for {
				timer := time.NewTimer(10 * time.Second)
				select {
				case req := <-reqch:
					if req == nil {
						return
					}
					if req.Type == "ping" {
						err := req.Reply(true, nil)
						switch err {
						case nil:
							timer.Stop()
						default:
							return
						}
					}
				case <-timer.C:
					log.Println("idle timeout")
					return
				case <-node.Closed():
					return
				}
			}
		})
		handleForward := func(ch ssh.NewChannel) {
			addr := string(ch.ExtraData())
			conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
			if err != nil {
				log.Println(err)
				return
			}
			defer conn.Close()
			tools.KeepAlive(conn, 5)
			sshChan, reqChan, err := ch.Accept()
			if err != nil {
				log.Println(err)
				return
			}
			defer sshChan.Close()
			done := make(chan interface{}, 2)
			go func() {
				for range reqChan {
				}
			}()
			go func() {
				io.Copy(sshChan, conn)
				done <- true
			}()
			go func() {
				io.Copy(conn, sshChan)
				done <- true
			}()
			select {
			case <-done: //close on first error
			case <-node.Closed():
			}
		}
		setupForward := func(ch ssh.NewChannel, cid string) {
			defer node.IfRecoverAction(func() {
				ch.Reject(ssh.ConnectionFailed, "closing")
			})
			child := node.AddChild(cid)
			child.AddProcess("handler", func() {
				defer child.IfRecoverAction(func() {
					ch.Reject(ssh.ConnectionFailed, "panic")
				})
				log.Println("open", count.increment(), cid)
				defer func() { log.Println("close", count.decrement(), cid) }()
				handleForward(ch)
			})
		}
		node.AddProcess("sshch", func() {
			for ch := range sshch {
				if ch.ChannelType() != "forward" {
					ch.Reject(ssh.Prohibited, "unsupported")
					return
				}
				addr := string(ch.ExtraData())
				cid := id.next(addr)
				setupForward(ch, cid)
			}
		})
		return
	}
	log.Panicln("connection failed", txts)
}
