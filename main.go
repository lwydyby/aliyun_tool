package main

import (
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	"aliyun/config"
	"aliyun/server"
)

//go:embed front/dist
var reactApp embed.FS

func main() {
	config.LoadYaml()
	// Set log flags.
	log.SetFlags(log.Lshortfile | log.Ltime)

	// Start server.
	server := server.NewSimpleServer(reactApp)
	server.Start()

	// Graceful shutdown.
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-stopChan
	log.Printf("received signal %s", sig)
	server.Stop()
}
