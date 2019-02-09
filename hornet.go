package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func parseArgs() string {
	path := flag.String("conf", "hornet.yaml", "conf file path")
	flag.Parse()
	return *path
}

func handleSignal(svr *Server) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-sigs
		Warn("get signal ", sig)

		switch sig {
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGINT:
			svr.Stop()
			break
		case syscall.SIGUSR2:
			InitLog()
		}
	}
}

func main() {
	path := parseArgs()

	LoadConf(path, path+".local")
	InitLog()

	Warn(GConfig)

	defer func() {
		if err := recover(); err != nil {
			Error(err)
		}
	}()

	store := new(Store)
	store.Init()

	svr := new(Server)
	svr.Init(store)
	go handleSignal(svr)

	svr.Forever()
	store.Close()
}