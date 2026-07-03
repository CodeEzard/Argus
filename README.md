# Argus

> AI-powered infrastructure monitor that detects anomalies and suggests fixes.

## Quickstart

```bash
# 1. Start Prometheus (if not already running)
docker run -p 9090:9090 prom/prometheus

# 2. Clone and run
git clone https://github.com/yourusername/argus
cd argus
go run . scan

# 3. Point at a different Prometheus host
go run . scan --host http://myserver:9090
```

## What it does right now (Phase 1)

- Connects to Prometheus
- Evaluates built-in rules (CPU usage, target health)
- Prints anomalies with severity to stdout
- Exits with code 1 if anomalies found (CI-friendly)

## Roadmap

- [ ] Phase 2: Z-score anomaly detection (replace dumb thresholds)
- [ ] Phase 3: `argus init` — auto-detect environment, write config
- [ ] Phase 4: AI-powered suggestions via Ollama / OpenAI
- [ ] Phase 5: Event history with SQLite
- [ ] Phase 6: `argus watch` — continuous monitoring daemon

## Project structure

```
argus/
├── main.go                      # binary entrypoint (tiny on purpose)
├── argus.yaml                   # config file
├── cmd/
│   ├── root.go                  # root cobra command + config init
│   └── scan.go                  # `argus scan` command
└── internal/
    ├── prometheus/
    │   └── client.go            # talks to Prometheus HTTP API
    └── detector/
        └── detector.go          # anomaly detection rules + logic
```
