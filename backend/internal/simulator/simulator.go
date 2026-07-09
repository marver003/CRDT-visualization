package simulator

import "net/http"

type Simulator struct {
    Registry *Registry
    Client   *http.Client
}

