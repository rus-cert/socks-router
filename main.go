package main

import (
	"flag"
	"net"
	"strings"
	"sync"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/rus-cert/socks-router/log"
	"github.com/rus-cert/socks-router/routing"
)

type stringList struct {
	list     []string
	defValue []string
}

func (i *stringList) String() string {
	return strings.Join(i.Get(), ",")
}

func (i *stringList) Set(value string) error {
	i.list = append(i.list, value)
	return nil
}

func (i *stringList) Get() []string {
	if 0 == len(i.list) {
		return i.defValue
	}
	return i.list
}

var configFile string
var listenAddrsVar = stringList{nil, []string{"127.0.0.1:8000", "[::1]:8000"}}
var debugFlag bool

func init() {
	defConfig, _ := homedir.Expand("~/.socks-routes")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	flag.StringVar(&configFile, "config", defConfig, "Path to configfile")
	flag.Var(&listenAddrsVar, "listen", "TCP Address to bind proxy to; can be passed multiple times")
}

func main() {
	flag.Parse()
	if debugFlag {
		log.EnableDebug()
	}
	listenAddrs := listenAddrsVar.Get()

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

		var wg sync.WaitGroup

		for _, addr := range listenAddrs {
			var listener *net.TCPListener
			log.Info.Printf("socks router listening on %v", addr)
			if l, err := net.Listen("tcp", addr); nil != err {
				log.Error.Fatalf("listen failed: %v", err)
			} else if l, ok := l.(*net.TCPListener); !ok {
				log.Error.Fatal("listen failed: not TCP")
			} else {
				listener = l
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer listener.Close()

				log.Error.Fatal(pm.ListenTCP(listener))
			}()
		}
		wg.Wait()
	}
}
