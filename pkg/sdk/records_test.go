package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/lanyi1998/DNSlog-GO/internal/model"
)

func TestRecordQueriesPreserveServerIPLocation(t *testing.T) {
	record := model.DnsInfo{Type: model.DNS, Subdomain: "abc.agent.example", Ipaddress: "127.0.0.1", IpLocation: "Shanghai", Request: "request data", Time: 123}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Code int             `json:"code"`
			Data []model.DnsInfo `json:"data"`
		}{Code: 200, Data: []model.DnsInfo{record}})
	}))
	defer server.Close()
	client := &DnsLogClient{httpClient: resty.New().SetBaseURL(server.URL)}
	for name, query := range map[string]func() ([]DnsRecord, error){"get": client.GetDnsData, "get-and-clear": client.GetDnsDataAndClear} {
		t.Run(name, func(t *testing.T) {
			records, err := query()
			if err != nil {
				t.Fatal(err)
			}
			if len(records) != 1 {
				t.Fatalf("records = %v", records)
			}
			want := DnsRecord{Type: string(record.Type), Subdomain: record.Subdomain, Ipaddress: record.Ipaddress, IpLocation: record.IpLocation, Request: record.Request, Time: record.Time}
			if records[0] != want {
				t.Fatalf("record = %+v, want %+v", records[0], want)
			}
			encoded, err := json.Marshal(records[0])
			if err != nil {
				t.Fatal(err)
			}
			var output map[string]any
			if err := json.Unmarshal(encoded, &output); err != nil {
				t.Fatal(err)
			}
			if output["ip_location"] != record.IpLocation {
				t.Fatalf("output = %s", encoded)
			}
			var roundTrip DnsRecord
			if err := json.Unmarshal(encoded, &roundTrip); err != nil {
				t.Fatal(err)
			}
			if roundTrip != want {
				t.Fatalf("round trip = %+v", roundTrip)
			}
		})
	}
}
