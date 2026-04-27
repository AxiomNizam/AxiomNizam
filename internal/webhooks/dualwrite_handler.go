package webhooks

import (
	"time"

	"example.com/axiomnizam/internal/platform/dualwrite"
	"example.com/axiomnizam/internal/platform/featureflags"
	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/resources"
)

const webhookDWModule = "webhooks"

type WebhookDualWriteStore = store.ResourceStore[*WebhookResource]

func (h *WebhookHandler) SetDualWriteStore(s WebhookDualWriteStore) { h.dualWriteStore = s }

func (h *WebhookHandler) isAuthoritative() bool {
	return h.dualWriteStore != nil && featureflags.ReconcilerAuthoritative(webhookDWModule)
}

func (h *WebhookHandler) buildWebhookResource(wh *Webhook) *WebhookResource {
	return &WebhookResource{
		TypeMeta:   resources.TypeMeta{APIVersion: WebhookAPIVersion, Kind: WebhookKind},
		ObjectMeta: resources.ObjectMeta{Name: wh.ID, Namespace: "default", Generation: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Spec:       WebhookSpec{TenantID: wh.TenantID, Description: wh.Description, URL: wh.URL, Secret: wh.Secret, Events: wh.Events, Active: wh.Active, Tags: wh.Tags},
		Status:     WebhookResourceStatus{ObjectStatus: resources.ObjectStatus{Phase: "Pending"}},
	}
}

func (h *WebhookHandler) dualWriteWebhook(wh *Webhook) {
	if h.dualWriteStore == nil || wh == nil {
		return
	}
	resource := h.buildWebhookResource(wh)
	resource.Status.Phase = "Ready"
	dualwrite.Write(webhookDWModule, h.dualWriteStore, resource)
}
