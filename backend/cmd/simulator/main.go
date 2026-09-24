package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/marver003/crdt/internal/config"
	"github.com/marver003/crdt/internal/simulator"
	simhttp "github.com/marver003/crdt/internal/transport/http/simulator"
)

func loadConfig(path string) (*config.Simulator, error) {
	if path == "" {
		return config.DefaultSimulator(), nil
	}

	return config.LoadSimulator(path)
}

func main() {
	configPath := flag.String("config", "", "Path to the simulator YAML config. Starts an empty simulator when omitted.")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	snapshots, err := cfg.Snapshots()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	sim, err := simulator.Load(cfg.Simulator.Crdt, snapshots)
	if err != nil {
		log.Fatalf("simulator: %v", err)
	}

	handler := simhttp.NewHandler(sim)
	router := simhttp.NewRouter(handler)

	log.Printf("Simulator (%s, %d nodes) API listening on %d", cfg.Simulator.Crdt, len(snapshots), cfg.Simulator.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Simulator.Port), router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
