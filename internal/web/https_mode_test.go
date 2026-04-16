package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"golang.org/x/crypto/acme/autocert"
)

func TestResolveHTTPSPlan(t *testing.T) {
	tests := []struct {
		name         string
		network      configs.Network
		filePaths    configs.FilePaths
		wantMode     httpsMode
		wantFallback string
	}{
		{
			name: "manual TLS takes precedence",
			network: configs.Network{
				HttpPort:  80,
				HttpsPort: 443,
			},
			filePaths: configs.FilePaths{
				WebDomain:     "play.example.com",
				HttpsCertFile: "cert.pem",
				HttpsKeyFile:  "key.pem",
			},
			wantMode: httpsModeManual,
		},
		{
			name: "auto TLS for public host on standard ports",
			network: configs.Network{
				HttpPort:  80,
				HttpsPort: 443,
			},
			filePaths: configs.FilePaths{
				WebDomain: "play.example.com",
			},
			wantMode: httpsModeAuto,
		},
		{
			name: "localhost stays on HTTP",
			network: configs.Network{
				HttpPort:  80,
				HttpsPort: 443,
			},
			filePaths: configs.FilePaths{
				WebDomain: "localhost",
			},
			wantMode:     httpsModeHTTPOnly,
			wantFallback: `automatic HTTPS requires a public hostname, got "localhost"`,
		},
		{
			name: "ip address stays on HTTP",
			network: configs.Network{
				HttpPort:  80,
				HttpsPort: 443,
			},
			filePaths: configs.FilePaths{
				WebDomain: "203.0.113.10",
			},
			wantMode:     httpsModeHTTPOnly,
			wantFallback: `automatic HTTPS requires a public hostname, got "203.0.113.10"`,
		},
		{
			name: "non standard ports disable auto TLS",
			network: configs.Network{
				HttpPort:  8080,
				HttpsPort: 8443,
			},
			filePaths: configs.FilePaths{
				WebDomain: "play.example.com",
			},
			wantMode:     httpsModeHTTPOnly,
			wantFallback: "automatic HTTPS requires Network.HttpPort=80 and Network.HttpsPort=443, got 8080/8443",
		},
		{
			name: "partial manual config falls back to HTTP",
			network: configs.Network{
				HttpPort:  80,
				HttpsPort: 443,
			},
			filePaths: configs.FilePaths{
				WebDomain:     "play.example.com",
				HttpsCertFile: "cert.pem",
			},
			wantMode:     httpsModeHTTPOnly,
			wantFallback: "manual HTTPS requires both HttpsCertFile and HttpsKeyFile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveHTTPSPlan(tt.network, tt.filePaths)
			if got.mode != tt.wantMode {
				t.Fatalf("resolveHTTPSPlan() mode = %q, want %q", got.mode, tt.wantMode)
			}
			if got.fallbackReason != tt.wantFallback {
				t.Fatalf("resolveHTTPSPlan() fallbackReason = %q, want %q", got.fallbackReason, tt.wantFallback)
			}
		})
	}
}

func TestBuildHTTPSRedirectTarget(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		httpsPort  int
		requestURI string
		want       string
	}{
		{
			name:       "default https port omits explicit port",
			host:       "play.example.com:80",
			httpsPort:  443,
			requestURI: "/webclient?x=1",
			want:       "https://play.example.com/webclient?x=1",
		},
		{
			name:       "non standard https port is preserved",
			host:       "play.example.com:8080",
			httpsPort:  8443,
			requestURI: "/admin/",
			want:       "https://play.example.com:8443/admin/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHTTPSRedirectTarget(tt.host, tt.httpsPort, tt.requestURI)
			if got != tt.want {
				t.Fatalf("buildHTTPSRedirectTarget() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultHTTPSGuidance(t *testing.T) {
	plan := httpsPlan{
		host:           "localhost",
		certFile:       "cert.pem",
		fallbackReason: "manual HTTPS requires both HttpsCertFile and HttpsKeyFile",
	}
	network := configs.Network{
		HttpPort:  8080,
		HttpsPort: 8443,
	}

	steps := defaultHTTPSGuidance(plan, network)
	if len(steps) == 0 {
		t.Fatalf("defaultHTTPSGuidance() returned no steps")
	}
}

func TestBuildAutoHTTPHandlerPassesThroughWhenRedirectDisabled(t *testing.T) {
	manager := &autocert.Manager{
		HostPolicy: autocert.HostWhitelist("play.example.com"),
	}
	network := configs.Network{
		HttpPort:      80,
		HttpsPort:     443,
		HttpsRedirect: false,
	}

	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler := buildAutoHTTPHandler(manager, network, fallback)
	req := httptest.NewRequest(http.MethodGet, "/webclient", nil)
	req.Host = "play.example.com"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("buildAutoHTTPHandler() status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestBuildAutoHTTPHandlerRedirectsWhenEnabled(t *testing.T) {
	manager := &autocert.Manager{
		HostPolicy: autocert.HostWhitelist("play.example.com"),
	}
	network := configs.Network{
		HttpPort:      80,
		HttpsPort:     443,
		HttpsRedirect: true,
	}

	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("fallback should not be used when redirect is enabled")
	})

	handler := buildAutoHTTPHandler(manager, network, fallback)
	req := httptest.NewRequest(http.MethodGet, "/webclient?x=1", nil)
	req.Host = "play.example.com"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("buildAutoHTTPHandler() status = %d, want %d", rec.Code, http.StatusMovedPermanently)
	}
	if got := rec.Header().Get("Location"); got != "https://play.example.com/webclient?x=1" {
		t.Fatalf("buildAutoHTTPHandler() Location = %q, want %q", got, "https://play.example.com/webclient?x=1")
	}
}

func TestBuildAutoHTTPHandlerInterceptsACMEChallenge(t *testing.T) {
	manager := &autocert.Manager{
		HostPolicy: autocert.HostWhitelist("play.example.com"),
	}
	network := configs.Network{
		HttpPort:      80,
		HttpsPort:     443,
		HttpsRedirect: false,
	}

	fallbackCalls := 0
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalls++
		w.WriteHeader(http.StatusNoContent)
	})

	handler := buildAutoHTTPHandler(manager, network, fallback)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/acme-challenge/token", nil)
	req.Host = "invalid.example"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if fallbackCalls != 0 {
		t.Fatalf("buildAutoHTTPHandler() called fallback for ACME challenge request")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("buildAutoHTTPHandler() status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
