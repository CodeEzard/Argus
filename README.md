# Argus 👁️

> You wake up to a 3am alert. Something's wrong.
> Grafana shows a spike. But what caused it? What do you do?
>
> **Argus tells you.**

Argus is an open-source, CLI-first infrastructure monitor that watches your metrics, detects anomalies using statistical analysis, correlates them across your stack, and uses a local LLM to tell you exactly what's wrong and how to fix it — in plain English, with exact commands.

No dashboards. No SaaS. No data leaving your machine.

```
───────────────────── 14:23:11 ─────────────────────

  ✓  CPU Usage          2.31%
  ✓  Load Average       0.18%
  ✗  Memory Usage       87.4%    [zscore + trend — z-score: 3.82]

  📋 Memory Usage is anomalous (value: 87.4, z-score: 3.82, severity: high)
     Triggered by: zscore, trend
     2 correlated metric(s):
       → Swap Usage      (likely_symptom, detected 12s ago)
       → Disk Writes/s   (likely_cause,   detected 34s ago)

  ┌─ 🔴 ANOMALY: Memory Usage ──────────────────────────────────────┐
  │  Severity     : HIGH                                            │
  │  Diagnosis    : Memory pressure detected. Disk write spike      │
  │                 preceded the memory anomaly by 34s, suggesting  │
  │                 a process is buffering large writes in RAM.      │
  │  Commands     :                                                  │
  │    $ ps aux --sort=-%mem | head -10                             │
  │    $ cat /proc/meminfo | grep -E 'Dirty|Writeback'             │
  │  Long term    : Add memory limits in docker-compose.yml,        │
  │                 consider tuning vm.dirty_ratio                  │
  │  Confidence   : 85%                                             │
  └─────────────────────────────────────────────────────────────────┘
```

---

## Why Argus

| Tool | Detects | Explains | Suggests fix | Self-hostable | Free |
|---|---|---|---|---|---|
| Grafana | ✅ | ❌ | ❌ | ✅ | ✅ |
| Datadog | ✅ | ⚠️ | ⚠️ | ❌ | ❌ |
| Dynatrace | ✅ | ✅ | ⚠️ | ❌ | ❌ |
| **Argus** | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## Features

- **Multi-signal anomaly detection** — z-score, trend analysis, and rate-of-change on every metric simultaneously
- **Cross-metric correlation** — finds which metrics spiked together and infers causal ordering by timestamp
- **Local LLM diagnosis** — runs against Ollama by default, no data leaves your machine
- **Natural language observations** — plain English summary before the LLM even runs
- **13 built-in metrics** — CPU, memory, swap, disk, network, load average, open files, and more
- **SQLite event store** — full history of every anomaly with `argus history` and `argus inspect`
- **Single binary** — no runtime dependencies, installs in seconds

---

## Quickstart

**Prerequisites:** Go 1.22+, Docker, Ollama

```bash
# 1. Clone
git clone https://github.com/CodeEzard/Argus
cd Argus

# 2. Start the monitoring stack
docker-compose up -d

# 3. Pull a local model
ollama pull llama3

# 4. Run
go run main.go watch --interval 15
```

That's it. Argus starts watching your infrastructure immediately.

---

## Commands

```bash
argus watch              # continuous monitoring (recommended)
argus watch --interval 5 # poll every 5 seconds
argus scan               # one-shot scan
argus history            # show all past anomalies
argus inspect <id>       # full detail on a specific anomaly
argus version            # print version
```

---

## How it works

```
Prometheus metrics
      ↓
Multi-signal detector
  ├── Z-score         (sudden spikes vs historical normal)
  ├── Trend           (gradual drift — memory leaks, disk filling)
  └── Rate of change  (acceleration — things getting worse fast)
      ↓
Cross-metric correlation
  (which metrics fired together? which came first?)
      ↓
Context snapshot
  (running services, system info, correlated metrics)
      ↓
Local LLM (Ollama)
  (diagnosis + exact commands + long-term fix)
      ↓
CLI output + SQLite storage
```

---

## Configuration

Argus reads from `argus.yaml` or environment variables:

```yaml
prometheus:
  host: http://localhost:9090   # default

store:
  path: argus.db                # default
```

Override via flags:
```bash
argus watch --host http://myserver:9090
```

Or via env:
```bash
ARGUS_PROMETHEUS_HOST=http://myserver:9090 argus watch
```

---

## Adding your own metrics

Edit `cmd/queries.go`:

```go
var queries = []Query{
    {Name: "My API latency", PromQL: `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))`},
    // ... add more
}
```

Any valid PromQL expression works.

---

## Roadmap

- [ ] `argus init` — auto-detect stack and configure exporters
- [ ] Webhook alerts (Slack, PagerDuty)
- [ ] Runbook generation and storage
- [ ] Web UI (optional, opt-in)
- [ ] Homebrew tap for one-command install

---

## License

MIT — see [LICENSE](LICENSE)

---

<p align="center">Built with Go · Runs on Prometheus · Powered by Ollama</p>