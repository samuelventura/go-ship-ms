package main

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/felixge/tcpkeepalive"
)

func keepAlive(conn net.Conn) error {
	return tcpkeepalive.SetKeepAlive(
		conn, 5*time.Second, 3, 1*time.Second)
}

func getenv(name string, defval string) string {
	value := os.Getenv(name)
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > 0 {
		log.Println(name, value)
		return value
	}
	log.Println(name, defval)
	return defval
}

func withext(ext string) string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Dir(exe)
	base := filepath.Base(exe)
	file := base + "." + ext
	return filepath.Join(dir, file)
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return hostname
}
