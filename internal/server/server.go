package server

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"l0/internal/store"

	"github.com/nats-io/nats.go"
)

type Server struct {
	Store          *store.Store
	Config         *Config
	NatsConnection *nats.Conn
}

func New(store *store.Store, config *Config) *Server {
	return &Server{
		Store:  store,
		Config: config,
	}
}

func (s *Server) Start() error {
	if err := s.configureStore(); err != nil {
		return err
	}

	if err := s.configureServer(); err != nil {
		return err
	}

	return nil
}

func (s *Server) configureStore() error {
	if err := s.Store.Open(); err != nil {
		return err
	}

	return nil
}

func (s *Server) configureServer() error {
	var err error

	s.NatsConnection, err = nats.Connect(s.Config.NatsURL)
	if err != nil {
		return err
	}

	_, err = s.NatsConnection.Subscribe("orders", s.OrdersHandler)
	if err != nil {
		return err
	}

	http.HandleFunc("/", s.GetOrderByIDHandler)

	go func() {
		fmt.Fprintln(os.Stderr, http.ListenAndServe(":8080", nil))
	}()

	return nil
}

func (s *Server) Stop() error {
	return s.Store.Close()
}
