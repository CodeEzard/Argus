# Argus

> An AI-powered SRE co-pilot and infrastructure monitor that detects anomalies in Prometheus metrics and generates structured, context-aware remediation suggestions using local LLMs.

---

## 🚀 Key Features

*   **Dynamic Anomaly Detection:** Employs statistical **Z-Score analysis** over a rolling window of metric values to dynamically establish baseline "normal" behavior, eliminating the need for hardcoded static thresholds.
*   **AI-Powered SRE Diagnostics:** Leverages a local **Ollama** integration (e.g., Llama 3.2) to analyze system snapshots, correlate simultaneous anomalies, diagnose root causes, and recommend step-by-step shell commands.
*   **Stateful Event Auditing:** Uses a lightweight, pure-Go **SQLite store** (zero-CGO) to log historical anomaly events, diagnostic suggestions, suggested fixes, and LLM confidence scores.
*   **CI/CD & DevOps Friendly:** Includes a one-shot `scan` command that exits with code `1` if any anomaly is detected, making it perfect for pipeline health-checks.
*   **Real-time Daemon (`watch`):** Continuously polls Prometheus at configurable intervals, evaluates metrics, runs detection logic, triggers LLM diagnostics, and persists reports.
*   **Cobra-Powered Interactive CLI:** Interactive commands to view metrics, list historical logs (`history`), and inspect specific events (`inspect <id>`).
*   **Viper Configuration Management:** Configuration via flags (`--host`), a YAML config file (`argus.yaml`), or environment variables prefixed with `ARGUS_`.

---

## 🛠️ Architecture

```
                 +-------------------+
                 |    Prometheus     |
                 +---------+---------+
                           |
                     (PromQL Poll)
                           v
                 +---------+---------+
                 |  Z-Score Detector |
                 +---------+---------+
                           |
                     (Anomaly Alert)
                           v
+--------------------------+--------------------------+
|                                                     |
|  +-------------------+      +--------------------+  |
|  |  System Collector |      |    Ollama LLM      |  |
|  |  (Docker Info, OS) | ---> |  (Llama 3.2:3b)    |  |
|  +-------------------+      +---------+----------+  |
|                                       |             |
|                                  (Diagnosis)        |
|                                       v             |
|  +-------------------+      +---------+----------+  |
|  |    SQLite DB      | <--- |   Console output   |  |
|  |  (argus.db Store) |      | (Formatted Table)  |  |
|  +-------------------+      +--------------------+  |
|                                                     |
|                     Argus Engine                    |
+-----------------------------------------------------+
```

---

## 📦 Tech Stack

*   **Language:** Go (Golang 1.25+)
*   **CLI Framework:** Cobra & Viper (for clean command composition and configuration hierarchy)
*   **Database:** SQLite (using `modernc.org/sqlite` for 100% CGO-free, cross-platform builds)
*   **LLM Provider:** Ollama (default: `llama3.2:3b`)
*   **Monitoring Source:** Prometheus HTTP API

---

## ⚙️ Quickstart

### 1. Run Prerequisites
Ensure you have a Prometheus instance and Ollama running:
```bash
# Start Prometheus
docker run -d -p 9090:9090 --name prometheus prom/prometheus

# Ensure Ollama is running Llama 3.2
ollama run llama3.2:3b
```

### 2. Install & Run
Clone the repository and build the binary:
```bash
# Clone the repository
git clone https://github.com/codeezard/argus.git
cd argus

# Build the project
go build -o argus .

# Run a one-shot metrics scan
./argus scan
```

---

## 💻 CLI Commands

### 🔍 Scan Metrics (`scan`)
Runs a one-shot metric check against the Prometheus host.
```bash
./argus scan --host http://localhost:9090
```

### 👁️ Real-time Monitoring (`watch`)
Start continuous background polling and trigger AI diagnostics on anomalies.
```bash
# Poll every 10 seconds (default is 30s)
./argus watch --interval 10
```

### 📋 View Event History (`history`)
Retrieve a structured table of all previously recorded anomalies from the local SQLite database.
```bash
./argus history
```

### 🔎 Detailed Inspection (`inspect`)
Show full details, LLM diagnosis, confidence level, and suggested remediation commands for a specific event ID.
```bash
./argus inspect 1
```

---

## 🎛️ Configuration

Argus looks for configurations in the following order:
1.  **Command-Line Flags** (e.g., `--host http://localhost:9090`)
2.  **Environment Variables** (prefixed with `ARGUS_`, e.g., `ARGUS_PROMETHEUS_HOST`)
3.  **Local Configuration File** (`./argus.yaml` or `$HOME/.argus/argus.yaml`)

**Example `argus.yaml`:**
```yaml
prometheus:
  host: "http://localhost:9090"
store:
  path: "argus.db"
```

