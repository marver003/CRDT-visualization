package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/marver003/crdt/internal/config"
	"github.com/marver003/crdt/internal/simulator"
	simhttp "github.com/marver003/crdt/internal/transport/http/simulator"
	"github.com/marver003/crdt/internal/types"
	"github.com/marver003/crdt/internal/web"
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

	newSim := func(crdtType types.CrdtType) (*simulator.Simulator, error) {
		if crdtType != cfg.Simulator.Crdt {
			return simulator.New(crdtType), nil
		}

		initial, err := cfg.Snapshots()
		if err != nil {
			return nil, err
		}

		return simulator.Load(cfg.Simulator.Crdt, initial)
	}

	sim, err := newSim(cfg.Simulator.Crdt)
	if err != nil {
		log.Fatalf("simulator: %v", err)
	}

	handler := simhttp.NewHandler(sim, newSim)
	router := simhttp.NewRouter(handler)

	log.Printf("Simulator (%s, %d nodes) API listening on %d", cfg.Simulator.Crdt, len(snapshots), cfg.Simulator.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.Simulator.Port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	if web.HasUI() {
		url := fmt.Sprintf("http://localhost:%d", cfg.Simulator.Port)
		log.Printf("Open %s in your browser", url)

		if err := web.OpenBrowser(url); err != nil {
			log.Printf("could not open the browser: %v", err)
		}
	}

	if err := http.Serve(listener, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
