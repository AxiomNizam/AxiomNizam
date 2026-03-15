package eventbus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupEventBusAckReplayRouter() (*gin.Engine, *InMemoryEventBusManager) {
	gin.SetMode(gin.TestMode)
	manager := NewInMemoryEventBusManager()
	handler := NewEventBusHandler(manager)

	router := gin.New()
	group := router.Group("/api/v1/eventbus")
	{
		group.POST("/events/publish", handler.PublishEvent)
		group.POST("/events/:id/ack", handler.AckEvent)
		group.POST("/dlq/:id/replay", handler.ReplayDLQEvent)
	}

	return router, manager
}

func performEventBusRequest(t *testing.T, router *gin.Engine, method, path string, payload interface{}) *httptest.ResponseRecorder {
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

func decodeEventBusBody(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response: %v; body=%s", err, rr.Body.String())
	}
}

func TestEventBusAckAndReplayEndpoints(t *testing.T) {
	router, manager := setupEventBusAckReplayRouter()

	t.Run("ack event success", func(t *testing.T) {
		_, _ = manager.CreateSubscription(&EventSubscription{
			ID:        "sub-1",
			TenantID:  "tenant-1",
			Name:      "orders-consumer",
			Topics:    []string{"orders.created"},
			Handler:   "worker://orders",
			CreatedAt: time.Now(),
		})

		publishResp := performEventBusRequest(t, router, http.MethodPost, "/api/v1/eventbus/events/publish", map[string]interface{}{
			"type":    "orders.created",
			"subject": "order-100",
			"source":  "orders-api",
			"data": map[string]interface{}{
				"orderId": "order-100",
			},
		})
		if publishResp.Code != http.StatusAccepted {
			t.Fatalf("expected publish 202, got %d: %s", publishResp.Code, publishResp.Body.String())
		}

		var publishBody EventPublishResponse
		decodeEventBusBody(t, publishResp, &publishBody)
		if publishBody.EventID == "" {
			t.Fatal("expected published event id")
		}

		ackResp := performEventBusRequest(t, router, http.MethodPost, "/api/v1/eventbus/events/"+publishBody.EventID+"/ack", map[string]interface{}{
			"subscriptionId": "sub-1",
			"acknowledgedBy": "worker-1",
			"message":        "processed successfully",
		})
		if ackResp.Code != http.StatusOK {
			t.Fatalf("expected ack 200, got %d: %s", ackResp.Code, ackResp.Body.String())
		}

		event, err := manager.GetEvent(publishBody.EventID)
		if err != nil {
			t.Fatalf("expected event to exist: %v", err)
		}
		if !event.IsProcessed {
			t.Fatal("expected acknowledged event to be marked processed")
		}
		if event.ProcessedAt.IsZero() {
			t.Fatal("expected processedAt to be set")
		}
		if event.Metadata["acknowledgedBy"] != "worker-1" {
			t.Fatalf("expected acknowledgedBy metadata, got %#v", event.Metadata["acknowledgedBy"])
		}
		if event.Metadata["subscriptionId"] != "sub-1" {
			t.Fatalf("expected subscriptionId metadata, got %#v", event.Metadata["subscriptionId"])
		}

		sub, err := manager.GetSubscription("sub-1")
		if err != nil {
			t.Fatalf("expected subscription to exist: %v", err)
		}
		if sub.ProcessedCount != 1 {
			t.Fatalf("expected subscription processed count to be 1, got %d", sub.ProcessedCount)
		}
	})

	t.Run("ack unknown event returns not found", func(t *testing.T) {
		ackResp := performEventBusRequest(t, router, http.MethodPost, "/api/v1/eventbus/events/missing-event/ack", map[string]interface{}{
			"acknowledgedBy": "worker-1",
		})
		if ackResp.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown event ack, got %d: %s", ackResp.Code, ackResp.Body.String())
		}
	})

	t.Run("replay dlq event success", func(t *testing.T) {
		baseEvent := EventBusEvent{
			ID:       "event-original-1",
			TenantID: "tenant-1",
			Type:     "orders.failed",
			Source:   "orders-worker",
			Data: map[string]interface{}{
				"orderId": "order-500",
			},
			Metadata: map[string]string{"attempt": "1"},
		}
		manager.events[baseEvent.ID] = &baseEvent
		manager.dlq["dlq-1"] = &DLQEvent{
			ID:              "dlq-1",
			OriginalEventID: baseEvent.ID,
			TenantID:        "tenant-1",
			Topic:           "orders.deadletter",
			FailureCount:    2,
			LastFailureTime: time.Now(),
			Event:           baseEvent,
		}

		replayResp := performEventBusRequest(t, router, http.MethodPost, "/api/v1/eventbus/dlq/dlq-1/replay", map[string]interface{}{
			"replayToTopic": "orders.replayed",
			"replayedBy":    "operator-1",
		})
		if replayResp.Code != http.StatusOK {
			t.Fatalf("expected replay 200, got %d: %s", replayResp.Code, replayResp.Body.String())
		}

		var replayBody struct {
			Replay EventPublishResponse `json:"replay"`
		}
		decodeEventBusBody(t, replayResp, &replayBody)
		if replayBody.Replay.EventID == "" {
			t.Fatal("expected replayed event id")
		}
		if replayBody.Replay.Topic != "orders.replayed" {
			t.Fatalf("expected replay topic orders.replayed, got %s", replayBody.Replay.Topic)
		}

		replayed, err := manager.GetEvent(replayBody.Replay.EventID)
		if err != nil {
			t.Fatalf("expected replayed event stored: %v", err)
		}
		if replayed.Metadata["replayedFromDLQ"] != "dlq-1" {
			t.Fatalf("expected replayedFromDLQ metadata, got %#v", replayed.Metadata["replayedFromDLQ"])
		}

		dlq := manager.dlq["dlq-1"]
		if !dlq.ManuallyResolved {
			t.Fatal("expected dlq event to be marked resolved")
		}
		if dlq.ResolutionAction != "retry" {
			t.Fatalf("expected resolutionAction=retry, got %s", dlq.ResolutionAction)
		}
		if dlq.ReplayToTopic != "orders.replayed" {
			t.Fatalf("expected replayToTopic updated, got %s", dlq.ReplayToTopic)
		}
	})

	t.Run("replay unknown dlq event returns not found", func(t *testing.T) {
		replayResp := performEventBusRequest(t, router, http.MethodPost, "/api/v1/eventbus/dlq/missing-dlq/replay", map[string]interface{}{
			"replayToTopic": "orders.replayed",
		})
		if replayResp.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown dlq replay, got %d: %s", replayResp.Code, replayResp.Body.String())
		}
	})
}
