package waitx

import "testing"

func TestParsePodReadinessDataList(t *testing.T) {
	raw := []byte(`{
  "items": [
    {
      "metadata": {"name": "api-0"},
      "status": {
        "phase": "Running",
        "conditions": [{"type": "Ready", "status": "True"}]
      }
    },
    {
      "metadata": {"name": "api-1"},
      "status": {
        "phase": "Running",
        "conditions": [{"type": "Ready", "status": "False"}]
      }
    }
  ]
}`)

	ready, total, names, err := parsePodReadinessData(raw)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if ready != 1 || total != 2 {
		t.Fatalf("unexpected readiness counts: ready=%d total=%d", ready, total)
	}
	if len(names) != 2 {
		t.Fatalf("unexpected pod name count: %d", len(names))
	}
}
