package cmd

var queries = []string{
	`100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])) * 100)`,
	`(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`,
}
