// Package store — tables.go
//
// Phase 5 of the etcd replacement plan: central registry of all
// resource table names.
//
// When using the Raft backend, all tables must be declared upfront
// so the shared go-memdb schema includes them all.  This file is the
// single source of truth for table names.
//
// Table names correspond to the etcd key prefix (without the
// /axiomnizam/ prefix and trailing slash).  For example, the etcd
// prefix "/axiomnizam/bulkoperations/" maps to table "bulkoperations".
package store

// AllResourceTables returns the complete list of resource table names
// used by the platform.  Each name maps 1:1 to an etcd prefix and a
// go-memdb table.
//
// When adding a new resource Kind, add its table name here.
func AllResourceTables() []string {
	return []string{
		// Internal KV table for direct key-value operations
		// (workflows, vectorplus, reviewflow, storage, IAM)
		"_kv",

		// Core reconciler resources (Phase 1 shadow mode)
		"bulkoperations",
		"eventbus-topics",
		"eventbus-subscriptions",
		"exportjobs",
		"streams",
		"rbac-roles",
		"rbac-rolebindings",
		"version-policies",
		"tracing-configs",
		"lineage-nodes",
		"audit-policies",
		"encryption-keys",
		"encryption-policies",
		"conductor-producers",
		"conductor-consumers",
		"webhooks",
		"tenants",
		"jobs",
		"etl-pipelines",
		"cdc-pipelines",
		"policies",
		"datasources",
		"iam-users",
		"api-scans",

		// Phase 6 resources
		"gis",
		"analytics-dashboards",
		"transform-rules",
		"notification-channels",
		"netintel-configs",

		// APIBanks
		"apibanks",

		// API Builder custom APIs
		"custom-apis",

		// WaitX health checks
		"wait-checks",

		// Platform Completion (WS-1 through WS-4)
		"catalogassets",
		"catalogcollections",
		"qualityrules",
		"qualitychecks",
		"schemas",
		"schemasubjects",
		"alertrules",
		"alertincidents",
		"notificationchannels",
	}
}
