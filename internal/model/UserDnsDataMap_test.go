package model

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func TestGetAndClearPreservesOtherTokensAndSubsequentWrites(t *testing.T) {
	data := NewDnsDataMap(nil)
	data.Set("token", DnsInfo{Subdomain: "old"})
	data.Set("other", DnsInfo{Subdomain: "other"})
	records := data.GetAndClear("token")
	if len(records) != 1 || records[0].Subdomain != "old" {
		t.Fatalf("records = %v", records)
	}
	if len(data.Get("token")) != 0 {
		t.Fatal("records were not cleared")
	}
	if len(data.Get("other")) != 1 {
		t.Fatal("another token was cleared")
	}
	data.Set("token", DnsInfo{Subdomain: "new"})
	if len(records) != 1 || records[0].Subdomain != "old" {
		t.Fatalf("returned records changed: %v", records)
	}
	if got := data.GetAndClear("token"); len(got) != 1 || got[0].Subdomain != "new" {
		t.Fatalf("subsequent records = %v", got)
	}
	if got := data.GetAndClear("token"); got == nil || len(got) != 0 {
		t.Fatalf("empty records = %v", got)
	}
}

func TestConcurrentWritesAndDrainsReturnEveryRecordOnce(t *testing.T) {
	const writers = 4
	const perWriter = 250
	data := NewDnsDataMap(nil)
	start := make(chan struct{})
	done := make(chan struct{})
	batches := make(chan []DnsInfo, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			var records []DnsInfo
			for {
				select {
				case <-done:
					batches <- records
					return
				default:
					records = append(records, data.GetAndClear("token")...)
					runtime.Gosched()
				}
			}
		}()
	}
	var pending sync.WaitGroup
	for writer := 0; writer < writers; writer++ {
		pending.Add(1)
		go func(id int) {
			defer pending.Done()
			<-start
			for i := 0; i < perWriter; i++ {
				data.Set("token", DnsInfo{Subdomain: fmt.Sprintf("%d-%d", id, i)})
			}
		}(writer)
	}
	close(start)
	pending.Wait()
	close(done)
	records := append(<-batches, (<-batches)...)
	records = append(records, data.GetAndClear("token")...)
	seen := make(map[string]bool)
	for _, record := range records {
		if seen[record.Subdomain] {
			t.Fatalf("duplicate record %q", record.Subdomain)
		}
		seen[record.Subdomain] = true
	}
	if len(seen) != writers*perWriter {
		t.Fatalf("received %d records, want %d", len(seen), writers*perWriter)
	}
}
