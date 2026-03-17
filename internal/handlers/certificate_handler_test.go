package handlers

import (
	"reflect"
	"testing"
)

const (
	testIDPHost   = "idp.bitbd.net"
	testIDPTarget = "idp.bitbd.net:443"
)

func TestNormalizeCertificateTarget(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTarget string
		wantHost   string
		wantPort   string
		wantErr    bool
	}{
		{
			name:       "host only defaults to 443",
			input:      testIDPHost,
			wantTarget: testIDPTarget,
			wantHost:   testIDPHost,
			wantPort:   "443",
		},
		{
			name:       "https URL with path",
			input:      "https://" + testIDPHost + "/admin/master/console/",
			wantTarget: testIDPTarget,
			wantHost:   testIDPHost,
			wantPort:   "443",
		},
		{
			name:       "host with explicit port",
			input:      "localhost:8443",
			wantTarget: "localhost:8443",
			wantHost:   "localhost",
			wantPort:   "8443",
		},
		{
			name:    "invalid port",
			input:   "localhost:abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTarget, gotHost, gotPort, err := normalizeCertificateTarget(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeCertificateTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotTarget != tt.wantTarget || gotHost != tt.wantHost || gotPort != tt.wantPort {
				t.Fatalf("normalizeCertificateTarget() = (%s, %s, %s), want (%s, %s, %s)", gotTarget, gotHost, gotPort, tt.wantTarget, tt.wantHost, tt.wantPort)
			}
		})
	}
}

func TestBuildRenewCommand(t *testing.T) {
	t.Run("replaces placeholder", func(t *testing.T) {
		got, err := buildRenewCommand("renew-cert --target {{target}} --force", testIDPTarget)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"renew-cert", "--target", testIDPTarget, "--force"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("buildRenewCommand() = %v, want %v", got, want)
		}
	})

	t.Run("appends target when placeholder is absent", func(t *testing.T) {
		got, err := buildRenewCommand("renew-cert --force", testIDPTarget)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"renew-cert", "--force", testIDPTarget}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("buildRenewCommand() = %v, want %v", got, want)
		}
	})
}
