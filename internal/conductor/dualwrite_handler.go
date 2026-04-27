package conductor

import (
	"time"

	"example.com/axiomnizam/internal/platform/dualwrite"
	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/resources"
)

const conductorDWModule = "conductor"

type ProducerDualWriteStore = store.ResourceStore[*ProducerResource]

func (h *Handler) SetProducerDualWriteStore(s ProducerDualWriteStore) { h.producerDualWriteStore = s }

func (h *Handler) dualWriteProducer(p *Producer) {
	if h.producerDualWriteStore == nil || p == nil {
		return
	}
	resource := &ProducerResource{
		TypeMeta:   resources.TypeMeta{APIVersion: ProducerAPIVersion, Kind: ProducerKind},
		ObjectMeta: resources.ObjectMeta{Name: p.ID, Namespace: "default", Generation: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Spec: ProducerSpec{
			Backend:     p.Backend,
			Exchange:    p.Exchange,
			RoutingKey:  p.RoutingKey,
			Topic:       p.Topic,
			ContentType: p.ContentType,
			Headers:     p.Headers,
			Config:      p.Config,
			Active:      p.Status == StatusActive,
		},
		Status: ProducerResourceStatus{ObjectStatus: resources.ObjectStatus{Phase: p.Status}, ProducerStatus: p.Status, MessagesSent: p.MessagesSent},
	}
	dualwrite.Write(conductorDWModule, h.producerDualWriteStore, resource)
}
