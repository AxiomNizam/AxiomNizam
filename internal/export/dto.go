package export

// ExportJobResponse is the API response for an export job.
type ExportJobResponse struct {
	ID            string       `json:"id"`
	Name          string       `json:"name,omitempty"`
	Status        ExportStatus `json:"status"`
	Format        string       `json:"format,omitempty"`
	Progress      float64      `json:"progress"`
	RecordCount   int64        `json:"recordCount,omitempty"`
	ProcessedRows int64        `json:"processedRows,omitempty"`
	SkippedRows   int64        `json:"skippedRows,omitempty"`
	ErrorRows     int64        `json:"errorRows,omitempty"`
	FileSize      int64        `json:"fileSize,omitempty"`
	CreatedAt     string       `json:"createdAt,omitempty"`
}

// ExportJobListResponse is the API response for listing export jobs.
type ExportJobListResponse struct {
	Exports []*ExportJob `json:"exports"`
	Count   int          `json:"count"`
}

// ProgressResponse is the API response for export progress.
type ProgressResponse struct {
	ID        string       `json:"id"`
	Status    ExportStatus `json:"status"`
	Progress  float64      `json:"progress"`
	Processed int64        `json:"processed"`
	Total     int64        `json:"total"`
	Skipped   int64        `json:"skipped"`
	Errors    int64        `json:"errors"`
}

// DownloadResponse is the API response for download URL.
type DownloadResponse struct {
	DownloadURL string `json:"downloadUrl"`
	FileSize    int64  `json:"fileSize"`
	ContentType string `json:"contentType"`
}

// TemplateListResponse is the API response for listing templates.
type TemplateListResponse struct {
	Templates []*ExportTemplate `json:"templates"`
	Count     int               `json:"count"`
}

// MessageResponse is a generic error response.
type MessageResponse struct {
	Error string `json:"error"`
}
