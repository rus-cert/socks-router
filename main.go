package main

import (
	"flag"
	"net"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/rus-cert/socks-router/log"
	"github.com/rus-cert/socks-router/routing"
)

var configFile string
var listenAddr string
var debugFlag bool

func init() {
	defConfig, _ := homedir.Expand("~/.socks-routes")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	flag.StringVar(&configFile, "config", defConfig, "Path to configfile")
	flag.StringVar(&listenAddr, "listen", "127.0.0.1:8000", "TCP Address to bind proxy to")
}

func main() {
	flag.Parse()
	if debugFlag {
		log.EnableDebug()
	}

	if routingMap, err := routing.ReadMapFile(configFile); nil != err {
		log.Error.Fatalf("Couldn't read config file: %v", err)
	} else {
		log.Info.Println("socks router starting")

		pm := ProtocolMultiplexer{}

		if socksHandler, err := CreateSocksHandler(routingMap); nil != err {
			log.Error.Fatal(err)
		} else {
			pm.Handlers = append(pm.Handlers, socksHandler)
		}

		if httpHandler, err := CreateHTTPHandler(routingMap); nil != err {
			log.Error.Fatal(err)
		} else {
			pm.Handlers = append(pm.Handlers, httpHandler)
		}

		var listener *net.TCPListener
		log.Info.Printf("socks router listening on %v", listenAddr)
		if l, err := net.Listen("tcp", listenAddr); nil != err {
			log.Error.Fatalf("listen failed: %v", err)
		} else if l, ok := l.(*net.TCPListener); !ok {
			log.Error.Fatal("listen failed: not TCP")
		} else {
			listener = l
		}
		defer listener.Close()

		log.Error.Fatal(pm.ListenTCP(listener))
	}
}
