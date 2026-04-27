package eventbus

import (
	"time"

	"example.com/axiomnizam/internal/platform/dualwrite"
	"example.com/axiomnizam/internal/platform/featureflags"
	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/resources"
)

const eventbusDWModule = "eventbus"

type TopicDualWriteStore = store.ResourceStore[*TopicResource]

func (h *EventBusHandler) SetTopicDualWriteStore(s TopicDualWriteStore) { h.topicDualWriteStore = s }

func (h *EventBusHandler) isAuthoritative() bool {
	return h.topicDualWriteStore != nil && featureflags.ReconcilerAuthoritative(eventbusDWModule)
}

func (h *EventBusHandler) buildTopicResource(topic *EventTopic) *TopicResource {
	return &TopicResource{
		TypeMeta:   resources.TypeMeta{APIVersion: TopicAPIVersion, Kind: TopicKind},
		ObjectMeta: resources.ObjectMeta{Name: topic.Name, Namespace: "default", Generation: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Spec:       TopicSpec{Description: topic.Description, Schema: topic.Schema, Partitions: topic.Partitions, ReplicationFactor: topic.ReplicationFactor, Retention: topic.Retention, Config: topic.Config, Active: topic.IsActive},
		Status:     TopicResourceStatus{ObjectStatus: resources.ObjectStatus{Phase: "Pending"}},
	}
}

func (h *EventBusHandler) dualWriteTopic(topic *EventTopic) {
	if h.topicDualWriteStore == nil || topic == nil {
		return
	}
	resource := h.buildTopicResource(topic)
	resource.Status.Phase = "Active"
	resource.Status.TopicActive = topic.IsActive
	dualwrite.Write(eventbusDWModule, h.topicDualWriteStore, resource)
}
