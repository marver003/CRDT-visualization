package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/marver003/crdt/internal/config"
)

const requestTimeout = 10 * time.Second

func main() {
	configPath := flag.String("config", "configs/livetest.yaml", "Path to the live test YAML config.")
	flag.Parse()

	cfg, err := config.LoadLiveTest(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	http.DefaultClient.Timeout = requestTimeout
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = cfg.Clients

	var wg sync.WaitGroup
	var mu sync.Mutex
	var slowest time.Duration

	start := time.Now()

	for c := 0; c < cfg.Clients; c++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			node := cfg.Nodes[c%len(cfg.Nodes)]

			for r := 0; r < cfg.Requests; r++ {
				t := time.Now()

				resp, err := http.Post(node+"/add", "application/json", nil)
				if err != nil {
					log.Fatalf("write to %s failed: %v", node, err)
				}

				resp.Body.Close()

				if resp.StatusCode != http.StatusNoContent {
					log.Fatalf("write to %s failed: HTTP %d", node, resp.StatusCode)
				}

				d := time.Since(t)

				mu.Lock()
				slowest = max(slowest, d)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	sending := time.Since(start)
	total := uint64(cfg.Clients * cfg.Requests)
	waitStart := time.Now()

	for !allEqual(values(cfg.Nodes), total) {
		if time.Since(waitStart) > cfg.Timeout {
			fmt.Printf("FAIL, expected %d on every node, got %v\n", total, values(cfg.Nodes))
			os.Exit(1)
		}

		time.Sleep(time.Millisecond)
	}

	fmt.Printf("OK, nodes=%d clients=%d requests=%d total=%d\n", len(cfg.Nodes), cfg.Clients, cfg.Requests, total)
	fmt.Println("sending:", sending)
	fmt.Println("slowest request:", slowest)
	fmt.Println("convergence:", time.Since(waitStart))
}

func values(nodes []string) []uint64 {
	result := make([]uint64, len(nodes))

	for i, node := range nodes {
		resp, err := http.Get(node + "/replicaState")
		if err != nil {
			log.Fatalf("read from %s failed: %v", node, err)
		}

		var state struct {
			Value uint64 `json:"value"`
		}

		json.NewDecoder(resp.Body).Decode(&state)
		resp.Body.Close()

		result[i] = state.Value
	}

	return result
}

func allEqual(values []uint64, total uint64) bool {
	for _, v := range values {
		if v != total {
			return false
		}
	}

	return true
}
