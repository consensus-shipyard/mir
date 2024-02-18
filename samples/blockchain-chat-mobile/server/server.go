package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	payloadpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	"github.com/rs/cors"
)

type Server struct {
	history            []*payloadpbtypes.Payload
	port               string
	newMessageCallback func(message string)
}

type postMsgPayload struct {
	Message string `json:"message"`
}

type getHistoryPayload struct {
	Messages []*payloadpbtypes.Payload `json:"messages"`
}

func (s *Server) handlePostMsg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed - Only POST", http.StatusMethodNotAllowed)
	}

	var pl postMsgPayload
	if err := json.NewDecoder(r.Body).Decode(&pl); err != nil {
		http.Error(w, "Expected {\"message\": \"some string\"}", http.StatusBadRequest)
	}

	// pass msg to application
	s.newMessageCallback(pl.Message)

	io.WriteString(w, "Submitted")
}

func (s *Server) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed - Only GET", http.StatusMethodNotAllowed)
	}

	json.NewEncoder(w).Encode(getHistoryPayload{s.history})
}

func (s *Server) SetHistory(history []*payloadpbtypes.Payload) {
	s.history = history
}

func (s *Server) Run() error {
	http.HandleFunc("/sendMessage", s.handlePostMsg)
	http.HandleFunc("/history", s.handleGetHistory)
	fmt.Printf("===== Server running on port %s\n", s.port)
	return http.ListenAndServe(s.port, cors.Default().Handler(http.DefaultServeMux))
}

func NewServer(port string, newMessageCallback func(message string)) *Server {
	return &Server{
		history:            []*payloadpbtypes.Payload{},
		port:               port,
		newMessageCallback: newMessageCallback,
	}
}
