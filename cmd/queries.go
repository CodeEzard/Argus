package cmd

type Query struct {
    Name   string
    PromQL string
}

var queries = []Query{
    // Compute
    {Name: "CPU Usage",      PromQL: `100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])) * 100)`},
    {Name: "CPU IOWait",     PromQL: `avg by(instance) (rate(node_cpu_seconds_total{mode="iowait"}[1m])) * 100`},
    {Name: "Load Average",   PromQL: `node_load1`},

    // Memory
    {Name: "Memory Usage",   PromQL: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`},
    {Name: "Swap Usage",     PromQL: `(1 - (node_memory_SwapFree_bytes / node_memory_SwapTotal_bytes)) * 100`},

    // Disk
    {Name: "Disk Usage",     PromQL: `100 - ((node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) * 100)`},
    {Name: "Disk Reads/s",   PromQL: `rate(node_disk_reads_completed_total[1m])`},
    {Name: "Disk Writes/s",  PromQL: `rate(node_disk_writes_completed_total[1m])`},

    // Network
    {Name: "Net Receive/s",  PromQL: `rate(node_network_receive_bytes_total{device!="lo"}[1m])`},
    {Name: "Net Transmit/s", PromQL: `rate(node_network_transmit_bytes_total{device!="lo"}[1m])`},
    {Name: "Net Errors",     PromQL: `rate(node_network_receive_errs_total[1m]) + rate(node_network_transmit_errs_total[1m])`},

    // Processes
    {Name: "Open Files",     PromQL: `node_filefd_allocated`},
    {Name: "Running Procs",  PromQL: `node_procs_running`},
}