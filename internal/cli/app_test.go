package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetDnsUsesConfiguredToken(t *testing.T) {
	const token = "test-token"

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/api/verifyToken":
			var body struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode verify token request: %v", err)
			}
			if body.Token != token {
				t.Errorf("verify token = %q", body.Token)
			}
			_, _ = writer.Write([]byte(`{"code":200,"msg":"success","data":{"subdomain":"agent.dns.example","token":"test-token"}}`))
		case "/api/getDnsData":
			if header := request.Header.Get("Token"); header != token {
				t.Errorf("Token header = %q", header)
			}
			_, _ = writer.Write([]byte(`{"code":200,"msg":"success","data":[{"Type":"DNS","Subdomain":"abc.agent.dns.example","Ipaddress":"127.0.0.1","Time":123}]}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	t.Setenv("DNSLOG_CONFIG", t.TempDir()+"/missing.yaml")
	t.Setenv("DNSLOG_BASE_URL", server.URL)
	t.Setenv("DNSLOG_TOKEN", token)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"get-dns"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"subdomain":"abc.agent.dns.example"`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if strings.Contains(stdout.String(), token) {
		t.Fatalf("stdout exposed token: %s", stdout.String())
	}
}

func TestUnknownCommandDoesNotRequireConfiguration(t *testing.T) {
	t.Setenv("DNSLOG_CONFIG", t.TempDir()+"/missing.yaml")
	t.Setenv("DNSLOG_BASE_URL", "")
	t.Setenv("DNSLOG_TOKEN", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"unknown"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestConfigSetWritesConfigWithoutExposingToken(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "dnslog-go", "config.yaml")
	t.Setenv("DNSLOG_CONFIG", configPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(
		[]string{"config", "set", "--url", "http://127.0.0.1:8000", "--token", "secret-token"},
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if strings.Contains(stdout.String(), "secret-token") {
		t.Fatalf("stdout exposed token: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), configPath) {
		t.Fatalf("stdout = %s", stdout.String())
	}

	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(contents), "secret-token") {
		t.Fatalf("config does not contain token: %s", contents)
	}
}

func TestVerificationCommandsPollUntilFound(t *testing.T) {
	testCases := []struct {
		name     string
		command  string
		endpoint string
		query    string
	}{
		{name: "DNS", command: "verify-dns", endpoint: "/api/verifyDns", query: "abc.agent.dns.example"},
		{name: "HTTP", command: "verify-http", endpoint: "/api/verifyHttp", query: "/agent/callback"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				switch request.URL.Path {
				case "/api/verifyToken":
					_, _ = writer.Write([]byte(`{"code":200,"msg":"success","data":{"subdomain":"agent.dns.example","token":"test-token"}}`))
				case testCase.endpoint:
					if attempts.Add(1) < 3 {
						_, _ = writer.Write([]byte(`{"code":200,"msg":"Not Found","data":null}`))
						return
					}
					_, _ = writer.Write([]byte(`{"code":200,"msg":"success","data":{"subdomain":"found","ipaddress":"127.0.0.1","time":123,"type":"DNS"}}`))
				default:
					http.NotFound(writer, request)
				}
			}))
			defer server.Close()

			t.Setenv("DNSLOG_CONFIG", t.TempDir()+"/missing.yaml")
			t.Setenv("DNSLOG_BASE_URL", server.URL)
			t.Setenv("DNSLOG_TOKEN", "test-token")

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(
				[]string{testCase.command, testCase.query, "--timeout", "1s", "--interval", "5ms"},
				&stdout,
				&stderr,
			)
			if exitCode != 0 {
				t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
			}
			if stdout.String() != "{\"found\":true}\n" {
				t.Fatalf("stdout = %s", stdout.String())
			}
			if attempts.Load() != 3 {
				t.Fatalf("verification attempts = %d", attempts.Load())
			}
		})
	}
}

func TestVerificationTimeoutZeroDisablesPolling(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/api/verifyToken":
			_, _ = writer.Write([]byte(`{"code":200,"msg":"success","data":{"subdomain":"agent.dns.example","token":"test-token"}}`))
		case "/api/verifyDns":
			attempts.Add(1)
			_, _ = writer.Write([]byte(`{"code":200,"msg":"Not Found","data":null}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	t.Setenv("DNSLOG_CONFIG", t.TempDir()+"/missing.yaml")
	t.Setenv("DNSLOG_BASE_URL", server.URL)
	t.Setenv("DNSLOG_TOKEN", "test-token")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"verify-dns", "missing.example", "--timeout", "0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if stdout.String() != "{\"found\":false}\n" {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if attempts.Load() != 1 {
		t.Fatalf("verification attempts = %d", attempts.Load())
	}
}

func TestVerificationDeadlineDistinguishesRequestFailureFromNoMatch(t *testing.T) {
	for _, command := range []string{"verify-dns", "verify-http"} {
		for _, slow := range []bool{false, true} {
			name := command + "/no-match"
			if slow {
				name = command + "/request-timeout"
			}
			t.Run(name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = io.Copy(io.Discard, r.Body)
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Path == "/api/verifyToken" {
						_, _ = w.Write([]byte(`{"code":200,"data":{"subdomain":"agent.example"}}`))
						return
					}
					if slow {
						select {
						case <-r.Context().Done():
						case <-time.After(time.Second):
							t.Error("timed-out request was not canceled")
						}
						return
					}
					_, _ = w.Write([]byte(`{"code":200,"msg":"Not Found","data":null}`))
				}))
				defer server.Close()
				t.Setenv("DNSLOG_CONFIG", t.TempDir()+"/missing.yaml")
				t.Setenv("DNSLOG_BASE_URL", server.URL)
				t.Setenv("DNSLOG_TOKEN", "test-token")
				var stdout, stderr bytes.Buffer
				code := Run([]string{command, "missing", "--timeout", "100ms", "--interval", "1s"}, &stdout, &stderr)
				if slow {
					if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "context deadline exceeded") {
						t.Fatalf("code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
					}
				} else if code != 0 || stdout.String() != "{\"found\":false}\n" || stderr.Len() != 0 {
					t.Fatalf("code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
				}
			})
		}
	}
}
