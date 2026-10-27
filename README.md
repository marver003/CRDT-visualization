# CRDT simulator

Simulator CRDT-jev (G-Counter na osnovi stanj in operacij, G-Set). Sestavljata ga strežnik v Go (`backend/`) in spletni vmesnik v TypeScriptu brez ogrodja (`frontend/`).

## Uporaba

1. Z GitHub Releases prenesite datoteko za svoj sistem:

   | Sistem | Datoteka |
   |---|---|
   | Windows | `simulator-windows-amd64.exe` |
   | macOS (Apple Silicon) | `simulator-darwin-arm64` |
   | macOS (Intel) | `simulator-darwin-amd64` |
   | Linux | `simulator-linux-amd64` |

2. Zaženite jo. Ob zagonu odpre brskalnik na `http://localhost:8080`.

   ```bash
   ./simulator-linux-amd64
   ```

3. Z zastavico `--config` se doda konfiguracijo z začetnim stanjem vozlišč (primeri so v `backend/configs/`):

   ```bash
   ./simulator-linux-amd64 --config simulator-gset.yaml
   ```

   Brez `--config` se zažene prazen G-Counter. Vrata določa polje `simulator.port` v konfiguraciji (privzeto 8080).

Opombe:

- macOS in Linux: datoteka mora imeti pravico izvajanja (`chmod +x ime-datoteke`).
- Strežnik posluša samo na `127.0.0.1`.

## Razvoj

Potrebujete Go in Node.js.

### Zagon strežnika

```bash
cd backend
go run ./cmd/simulator --config configs/simulator-gcounter.yaml
```

Konfiguracije so v `backend/configs/`: `simulator-gcounter.yaml`, `simulator-gset.yaml`, `simulator-gcounter-op.yaml`. 

### Vmesnik

Vmesnik je v `frontend/`, prevaja se 1:1 v `frontend/dist/` (brez bundlerja):

```bash
cd frontend
npm install
node node_modules/typescript/lib/tsc.js -p tsconfig.json
```

Nato odprite `frontend/index.html` neposredno v brskalniku. Brez dodatnih nastavitev se poveže na `http://localhost:8080`. Za drugačen naslov strežnika dodajte parameter:

```
index.html?api=http://localhost:9000
```

### Struktura

```
backend/
  cmd/simulator/        zagon simulatorja
  cmd/node/             produkcijska replika (gRPC)
  configs/              primeri konfiguracij
  internal/crdt/        G-Counter na osnovi stanja in operacij, G-Set
  internal/simulator/   dogodkovni model (koraki, sporočila, operacije)
  internal/transport/   HTTP API
  internal/web/         vgrajeni vmesnik
frontend/               TypeScript, HTML, CSS
```
