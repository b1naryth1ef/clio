package clerb

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Server struct {
	Router *http.ServeMux
	Server *http.Server
}

func NewServer(addr string) Server {
	// router := mux.NewRouter()
	router := http.NewServeMux()

	return Server{
		Router: router,
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Server) BindAll() {
	r := mux.NewRouter()

	// API Methods
	r.HandleFunc("/api/init", s.HandleAPIInit)
	s.Router.Handle("/", r)
}

func (s *Server) Serve() {
	s.Server.ListenAndServe()
}

func (s *Server) HandleAPIInit(w http.ResponseWriter, r *http.Request) {
	log.Printf("Aww yiss")
}
