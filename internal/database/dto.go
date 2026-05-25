package database

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// StatusResponse is a generic success response with a message.
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CreateDatabaseResponse is returned after successfully creating a database.
type CreateDatabaseResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Database   string `json:"database"`
	DBType     string `json:"db_type"`
	DBServer   string `json:"db_server"`
	ServerName string `json:"server_name"`
}

// ConnectDatabaseServerResponse is returned after connecting or updating a database server.
type ConnectDatabaseServerResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Server  DatabaseServerInfo `json:"server"`
}

// ListDatabaseServersResponse is returned by ListDatabaseServers.
type ListDatabaseServersResponse struct {
	Status  string               `json:"status"`
	Count   int                  `json:"count"`
	Servers []DatabaseServerInfo `json:"servers"`
}

// CreateTableResponse is returned after successfully creating a table.
type CreateTableResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Table   string `json:"table"`
	DBType  string `json:"db_type"`
	Columns int    `json:"columns"`
}

// ListDatabasesResponse is returned by ListDatabases.
type ListDatabasesResponse struct {
	Status    string   `json:"status"`
	DBType    string   `json:"db_type"`
	Databases []string `json:"databases"`
	Count     int      `json:"count"`
}

// ListTablesResponse is returned by ListTables.
type ListTablesResponse struct {
	Status string   `json:"status"`
	DBType string   `json:"db_type"`
	Tables []string `json:"tables"`
	Count  int      `json:"count"`
}
