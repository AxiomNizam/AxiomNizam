package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupJobSchedulingTestRouter(t *testing.T) (*gin.Engine, *JobHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewJobHandler()
	t.Cleanup(func() {
		handler.Close()
	})

	router := gin.New()
	router.POST("/api/v1/jobs", handler.Create)
	router.GET("/api/v1/jobs", handler.List)
	router.GET("/api/v1/jobs/schedules", handler.ListSchedules)
	router.GET("/api/v1/jobs/:id", handler.Get)
	router.POST("/api/v1/jobs/:id/schedule", handler.SetSchedule)
	router.DELETE("/api/v1/jobs/:id/schedule", handler.RemoveSchedule)
	router.POST("/api/v1/jobs/:id/run", handler.Run)
	return router, handler
}

func performJobSchedulingRequest(t *testing.T, router *gin.Engine, method, path string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		body = bytes.NewReader(encoded)
	} else {
		body = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func decodeJobSchedulingResponse(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response: %v; body=%s", err, rr.Body.String())
	}
}

func TestJobSchedulingAPIs(t *testing.T) {
	t.Run("create with schedule and list schedules", func(t *testing.T) {
		router, _ := setupJobSchedulingTestRouter(t)

		createResp := performJobSchedulingRequest(t, router, http.MethodPost, "/api/v1/jobs", map[string]interface{}{
			"metadata": map[string]interface{}{"name": "nightly-sync"},
			"spec": map[string]interface{}{
				"type":     "sync",
				"schedule": "1h",
			},
		})
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected create 201, got %d: %s", createResp.Code, createResp.Body.String())
		}

		var created JobResource
		decodeJobSchedulingResponse(t, createResp, &created)
		if created.Schedule == nil {
			t.Fatal("expected schedule metadata on created job")
		}
		if created.Schedule.Expression != "1h" {
			t.Fatalf("expected schedule expression 1h, got %q", created.Schedule.Expression)
		}
		if created.Schedule.NextRun == "" {
			t.Fatal("expected schedule nextRun to be set")
		}

		schedulesResp := performJobSchedulingRequest(t, router, http.MethodGet, "/api/v1/jobs/schedules", nil)
		if schedulesResp.Code != http.StatusOK {
			t.Fatalf("expected schedules 200, got %d: %s", schedulesResp.Code, schedulesResp.Body.String())
		}

		var listed struct {
			Schedules []map[string]interface{} `json:"schedules"`
			Count     int                      `json:"count"`
		}
		decodeJobSchedulingResponse(t, schedulesResp, &listed)
		if listed.Count != 1 || len(listed.Schedules) != 1 {
			t.Fatalf("expected one schedule, got count=%d body=%s", listed.Count, schedulesResp.Body.String())
		}
		if got, ok := listed.Schedules[0]["expression"].(string); !ok || got != "1h" {
			t.Fatalf("expected listed expression 1h, got %#v", listed.Schedules[0]["expression"])
		}
	})

	t.Run("set and remove schedule", func(t *testing.T) {
		router, _ := setupJobSchedulingTestRouter(t)

		createResp := performJobSchedulingRequest(t, router, http.MethodPost, "/api/v1/jobs", map[string]interface{}{
			"metadata": map[string]interface{}{"name": "manual-job"},
			"spec":     map[string]interface{}{"type": "backup"},
		})
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected create 201, got %d: %s", createResp.Code, createResp.Body.String())
		}

		var created JobResource
		decodeJobSchedulingResponse(t, createResp, &created)

		setResp := performJobSchedulingRequest(t, router, http.MethodPost, "/api/v1/jobs/"+created.Metadata.ID+"/schedule", map[string]interface{}{
			"schedule": "30m",
		})
		if setResp.Code != http.StatusOK {
			t.Fatalf("expected set schedule 200, got %d: %s", setResp.Code, setResp.Body.String())
		}

		var setBody struct {
			Job JobResource `json:"job"`
		}
		decodeJobSchedulingResponse(t, setResp, &setBody)
		if setBody.Job.Schedule == nil || setBody.Job.Schedule.Expression != "30m" {
			t.Fatalf("expected job schedule to be set to 30m, got %#v", setBody.Job.Schedule)
		}

		removeResp := performJobSchedulingRequest(t, router, http.MethodDelete, "/api/v1/jobs/"+created.Metadata.ID+"/schedule", nil)
		if removeResp.Code != http.StatusOK {
			t.Fatalf("expected remove schedule 200, got %d: %s", removeResp.Code, removeResp.Body.String())
		}

		getResp := performJobSchedulingRequest(t, router, http.MethodGet, "/api/v1/jobs/"+created.Metadata.ID, nil)
		if getResp.Code != http.StatusOK {
			t.Fatalf("expected get 200, got %d: %s", getResp.Code, getResp.Body.String())
		}
		var gotJob JobResource
		decodeJobSchedulingResponse(t, getResp, &gotJob)
		if gotJob.Schedule != nil {
			t.Fatalf("expected schedule removed, got %#v", gotJob.Schedule)
		}
	})

	t.Run("scheduled job trigger executes job", func(t *testing.T) {
		router, handler := setupJobSchedulingTestRouter(t)

		createResp := performJobSchedulingRequest(t, router, http.MethodPost, "/api/v1/jobs", map[string]interface{}{
			"metadata": map[string]interface{}{"name": "fast-scheduled"},
			"spec": map[string]interface{}{
				"type":     "sync",
				"schedule": "1s",
			},
		})
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected create 201, got %d: %s", createResp.Code, createResp.Body.String())
		}

		var created JobResource
		decodeJobSchedulingResponse(t, createResp, &created)
		if created.Schedule == nil {
			t.Fatal("expected schedule metadata")
		}

		handler.mu.Lock()
		job := handler.findJob(created.Metadata.ID)
		if job == nil {
			handler.mu.Unlock()
			t.Fatalf("job %s not found in handler state", created.Metadata.ID)
		}
		job.Schedule.NextRun = time.Now().Add(-2 * time.Second).UTC().Format(time.RFC3339)
		handler.mu.Unlock()

		handler.processScheduledJobs()

		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			getResp := performJobSchedulingRequest(t, router, http.MethodGet, "/api/v1/jobs/"+created.Metadata.ID, nil)
			if getResp.Code != http.StatusOK {
				t.Fatalf("expected get 200, got %d: %s", getResp.Code, getResp.Body.String())
			}
			var current JobResource
			decodeJobSchedulingResponse(t, getResp, &current)
			if current.Status.Phase == "Succeeded" {
				if current.Schedule == nil || current.Schedule.NextRun == "" {
					t.Fatal("expected next scheduled run to be set after execution")
				}
				return
			}
			time.Sleep(150 * time.Millisecond)
		}

		getResp := performJobSchedulingRequest(t, router, http.MethodGet, "/api/v1/jobs/"+created.Metadata.ID, nil)
		var final JobResource
		decodeJobSchedulingResponse(t, getResp, &final)
		t.Fatalf("expected scheduled execution to complete, final phase=%s, progress=%d, logs=%v", final.Status.Phase, final.Status.Progress, final.Logs)
	})
}

func TestJobSchedulingRejectsInvalidExpression(t *testing.T) {
	router, _ := setupJobSchedulingTestRouter(t)

	resp := performJobSchedulingRequest(t, router, http.MethodPost, "/api/v1/jobs", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "bad-schedule"},
		"spec": map[string]interface{}{
			"type":     "sync",
			"schedule": "not-a-schedule",
		},
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected create 400 for invalid schedule, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]interface{}
	decodeJobSchedulingResponse(t, resp, &body)
	if body["error"] == nil {
		t.Fatalf("expected error payload, got %v", body)
	}
}

func TestSetScheduleRequiresExpression(t *testing.T) {
	router, _ := setupJobSchedulingTestRouter(t)

	createResp := performJobSchedulingRequest(t, router, http.MethodPost, "/api/v1/jobs", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "missing-expression"},
		"spec":     map[string]interface{}{"type": "sync"},
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d: %s", createResp.Code, createResp.Body.String())
	}

	var created JobResource
	decodeJobSchedulingResponse(t, createResp, &created)

	setResp := performJobSchedulingRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/jobs/%s/schedule", created.Metadata.ID), map[string]interface{}{})
	if setResp.Code != http.StatusBadRequest {
		t.Fatalf("expected set schedule 400 when expression missing, got %d: %s", setResp.Code, setResp.Body.String())
	}
}
