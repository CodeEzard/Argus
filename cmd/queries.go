package cmd

var queries = []string{
    `rate(process_cpu_seconds_total[1m])`,
    `rate(prometheus_target_scrape_pool_exceeded_target_limit_total[5m])`,
}