package cmd

type Query struct {
    Name   string
    PromQL string
}

var queries = []Query{
    {
        Name:   "CPU Usage",
        PromQL: `100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])) * 100)`,
    },
    {
        Name:   "Memory Usage",
        PromQL: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`,
    },
}