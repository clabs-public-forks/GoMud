package web

import "testing"

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
		{
			name:       "ipv6 host with source port keeps brackets",
			host:       "[::1]:80",
			httpsPort:  443,
			requestURI: "/webclient",
			want:       "https://[::1]/webclient",
		},
		{
			name:       "ipv6 host with non standard https port keeps brackets",
			host:       "[2001:db8::10]:8080",
			httpsPort:  8443,
			requestURI: "/admin/",
			want:       "https://[2001:db8::10]:8443/admin/",
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
