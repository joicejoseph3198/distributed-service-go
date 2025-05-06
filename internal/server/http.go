package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

/*
When building a JSON/HTTP Go server,
each handler consists of three steps:
1. Unmarshal the request’s JSON body into a struct.
2. Run that endpoint’s logic with the request to obtain a result.
3. Marshal and write that result to the response.
If your handlers become much more complicated than this, then you should
move the code out, move request and response handling into HTTP middle-
ware, and move business logic further down the stack.
*/

type httpServer struct{
	Log *Log
}

// server referencing a log for the server to defer to in its handlers
func newHTTPServer() *httpServer {
	return &httpServer{
		Log: NewLog(),
	}
}

type ProduceRequest struct {
	Record Record `json:"record"`
}

type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type ConsumeResponse struct {
	Record Record `json:"record"`
}

func (s *httpServer) handleProduce(w http.ResponseWriter, r *http.Request){
	req := ProduceRequest{}
	// Creates a new JSON decoder that reads from the request body (r.Body).
	// Decodes the JSON from the request body into the req struct.
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil{
		http.Error(w,err.Error(), http.StatusBadRequest)
		return
	}
	off,err := s.Log.Append(req.Record)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	
	// encoding that struct into JSON, and writing it directly to the HTTP response
	res := ProduceResponse{Offset: off}
	err = json.NewEncoder(w).Encode(res)
	if err != nil{
		http.Error(w, err.Error(),http.StatusInternalServerError)
	}
}

func (s *httpServer) handleConsume(w http.ResponseWriter, r *http.Request){
	req := ConsumeRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := s.Log.Read(req.Offset)
	if err == ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ConsumeResponse{Record: record}
	err = json.NewEncoder(w).Encode(res)
	if err != nil{
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}


// http.Request is a struct, struct can be large, and passing it as a pointer:
// Avoids copying the entire struct. Allows the function to access/update internal fields
// http.ResponseWriter is an interface - they are reference types by design.
// When you receive w http.ResponseWriter, you're receiving a reference to a value that implements the interface.
func NewHTTPServer(addr string) *http.Server {
	server := newHTTPServer()
	r := mux.NewRouter()
	r.HandleFunc("/",server.handleProduce).Methods("POST")
	r.HandleFunc("/",server.handleConsume).Methods("GET")
	return &http.Server{
		Addr: addr,
		Handler: r,
	}
}

