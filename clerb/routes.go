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

	api_r := r.PathPrefix("/api").Subrouter()
	api_r.HandleFunc("/init", s.HandleAPIInit)

	// Bind our router to the HTTP router
	s.Router.Handle("/", r)
}

func (s *Server) Serve() {
	s.Server.ListenAndServe()
}

func (s *Server) HandleAPIInit(w http.ResponseWriter, r *http.Request) {
	log.Printf("Aww yiss")
}
