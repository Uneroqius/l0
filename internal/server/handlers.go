package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"l0/internal/order"

	"github.com/nats-io/nats.go"
)

func (s *Server) OrdersHandler(m *nats.Msg) {
	o := &order.Order{}
	if err := json.NewDecoder(bytes.NewBuffer(m.Data)).Decode(o); err != nil {
		fmt.Fprintln(os.Stderr, err) //недопустимые данные отбрасываются
		return
	}

	if err := s.Store.AddOrderToBD(o); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func (s *Server) GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	order_uid := r.URL.Query().Get("order_uid")

	if order_uid == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(fmt.Sprintf(`{"Error":"%s"}`, "input a order_uid")))
		return
	}

	order, err := s.Store.GetOrderByID(order_uid)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(fmt.Sprintf(`{"Error":"%s"}`, err.Error())))
		return
	}

	bufOrder := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(bufOrder).Encode(&order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(bufOrder.Bytes())
}
