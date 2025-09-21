package api

import (
    "context"
    "net/http"
    "time"
    
    "github.com/gorilla/mux"
    "github.com/Inkedup1114/dvpn/internal/api/handlers"
    "github.com/Inkedup1114/dvpn/pkg/mining"
    "github.com/Inkedup1114/dvpn/pkg/p2p"
)

type Server struct {
    router       *mux.Router
    httpServer   *http.Server
    statusHandler *handlers.StatusHandler
    miningHandler *handlers.MiningHandler
    peersHandler  *handlers.PeersHandler
}

func NewServer(addr string, miner mining.Miner, network *p2p.P2PNetwork) *Server {
    s := &Server{
        router:        mux.NewRouter(),
        statusHandler: handlers.NewStatusHandler(network),
        miningHandler: handlers.NewMiningHandler(miner),
        peersHandler:  handlers.NewPeersHandler(network),
    }
    
    s.setupRoutes()
    
    s.httpServer = &http.Server{
        Addr:         addr,
        Handler:      s.router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }
    
    return s
}

func (s *Server) setupRoutes() {
    api := s.router.PathPrefix("/api").Subrouter()
    
    // Health check
    api.HandleFunc("/health", s.statusHandler.GetHealth).Methods("GET")
    
    // Status endpoints
    api.HandleFunc("/status", s.statusHandler.GetStatus).Methods("GET")
    
    // Mining endpoints
    api.HandleFunc("/mining/start", s.miningHandler.StartMining).Methods("POST")
    api.HandleFunc("/mining/stop", s.miningHandler.StopMining).Methods("POST")
    api.HandleFunc("/mining/status", s.miningHandler.GetMiningStatus).Methods("GET")

    // Peer endpoints
    api.HandleFunc("/peers/connect", s.peersHandler.ConnectPeer).Methods("POST")
    api.HandleFunc("/peers", s.peersHandler.GetPeers).Methods("GET")
    
    // Serve static files (for future web UI)
    s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))
}

func (s *Server) Start() error {
    return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
    return s.httpServer.Shutdown(ctx)
}
