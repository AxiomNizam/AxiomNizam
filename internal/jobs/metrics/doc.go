// Package metrics provides Prometheus metric constants and labels for the jobs module.
//
// The actual Prometheus counters, gauges, and histograms are defined in the parent
// jobs package (jobs.MetricsCollector). This subpackage provides consistent label
// constants and value sets following the gatekeeper metrics pattern.
//
// For the full MetricsCollector, use jobs.NewMetricsCollector(namespace).
package metrics
