package main

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/samuelventura/go-tree"
	"golang.org/x/crypto/ssh"
)

func run(node tree.Node) {
	name := node.GetValue("name").(string)
	pool := node.GetValue("pool").(string)
	record := node.GetValue("record").(string)
	keypath := node.GetValue("keypath").(string)
	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		log.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
	} else {
		txts = []string{pool}
	}
	rand.Seed(time.Now().UnixNano())
	for _, txt := range txts {
		addrs := strings.Split(txt, ",")
		l := len(addrs)
		//random start
		n := rand.Intn(l)
		for i := 0; i < l; i++ {
			addr := addrs[(n+i)%l]
			log.Println(addr, name)
			conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
			if err != nil {
				log.Println(err)
				continue
			}
			err = keepAlive(conn)
			if err != nil {
				log.Println(err)
				conn.Close()
				continue
			}
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
				log.Println("open", addr)
				defer log.Println("close", addr)
				sshch, _, err := ch.Accept()
				if err != nil {
					log.Println(err)
					return
				}
				defer sshch.Close()
				conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
				if err != nil {
					log.Println(err)
					return
				}
				defer conn.Close()
				err = keepAlive(conn)
				if err != nil {
					log.Println(err)
					return
				}
				done := make(chan interface{}, 2)
				go func() {
					io.Copy(sshch, conn)
					done <- true
				}()
				go func() {
					io.Copy(conn, sshch)
					done <- true
				}()
				select {
				case <-done: //close on first error
				case <-node.Closed():
				}
			}
			node.AddProcess("sshch", func() {
				defer log.Println("channel handler exited")
				for ch := range sshch {
					if ch.ChannelType() != "forward" {
						ch.Reject(ssh.Prohibited, "unsupported")
						return
					}
					go handleForward(ch)
				}
			})
			return
		}
	}
	log.Fatalln("connection failed", txts)
}
