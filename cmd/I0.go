package main

import (
	"fmt"
	"os"

	"l0/internal/cl"
	"l0/internal/server"
	"l0/internal/store"
)

var (
	psqlURL string = "host=localhost user=nw password=123 dbname=t0 sslmode=disable"
	natsURL string = ":8080"
)

func main() {
	storeConfig := store.NewConfig(psqlURL)
	store := store.New(storeConfig)
	serverConfig := server.NewConfig(natsURL, storeConfig)
	server := server.New(store, serverConfig)

	if err := server.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	cl := cl.New(store)
	cl.Start()
}
