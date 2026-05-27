package audit

// Severity represents the severity of an audit event.
type Severity string

const (
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

// Category represents the category of an audit event.
type Category string

const (
	CategoryParser    Category = "parser"
	CategoryIngest    Category = "ingest"
	CategoryTopology  Category = "topology"
	CategoryAnomaly   Category = "anomaly"
	CategoryAlert     Category = "alert"
	CategoryMode      Category = "mode"
	CategoryLifecycle Category = "lifecycle"
)

// Action represents the action of an audit event.
type Action string

const (
	ActionParserCreated   Action = "parser_created"
	ActionParserUpdated   Action = "parser_updated"
	ActionParserDeleted   Action = "parser_deleted"
	ActionEntryIngested   Action = "entry_ingested"
	ActionEntryDropped    Action = "entry_dropped"
	ActionTopologyUpdated Action = "topology_updated"
	ActionAnomalyAcked    Action = "anomaly_acknowledged"
	ActionAnomalyResolved Action = "anomaly_resolved"
	ActionAlertAcked      Action = "alert_acknowledged"
	ActionAlertResolved   Action = "alert_resolved"
	ActionModeToggled     Action = "mode_toggled"
	ActionModuleStarted   Action = "module_started"
	ActionModuleStopped   Action = "module_stopped"
)
