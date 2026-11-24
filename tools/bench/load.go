// load runs concurrent uploads or concurrent reads against a kubin server.
// See docs/capacity.md.
//
//	go run load.go upload <base> <file.tar.gz> <concurrency> <total>
//	go run load.go read   <base> <snapshotID> <concurrency> <seconds> <ns/pod/container>
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func pct(d []time.Duration, p float64) time.Duration {
	if len(d) == 0 {
		return 0
	}
	sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
	return d[int(float64(len(d)-1)*p)]
}

func main() {
	if len(os.Args) < 6 || (os.Args[1] == "read" && len(os.Args) < 7) {
		fmt.Fprintln(os.Stderr, "usage:\n  load upload <base> <file.tar.gz> <concurrency> <total>\n  load read <base> <snapshotID> <concurrency> <seconds> <ns/pod/container>")
		os.Exit(2)
	}
	mode, base := os.Args[1], os.Args[2]
	conc, _ := strconv.Atoi(os.Args[4])
	client := &http.Client{Timeout: 10 * time.Minute, Transport: &http.Transport{MaxIdleConnsPerHost: 256}}
	var mu sync.Mutex
	var lat []time.Duration
	var errs, n atomic.Int64
	codes := map[int]int{}
	rec := func(d time.Duration, code int) {
		mu.Lock()
		lat = append(lat, d)
		codes[code]++
		mu.Unlock()
	}
	start := time.Now()
	var wg sync.WaitGroup
	switch mode {
	case "upload":
		body, _ := os.ReadFile(os.Args[3])
		total, _ := strconv.Atoi(os.Args[5])
		var next atomic.Int64
		var lastID atomic.Value
		for range conc {
			wg.Go(func() {
				for next.Add(1) <= int64(total) {
					t := time.Now()
					resp, err := client.Post(base+"/api/v1/snapshots", "application/gzip", bytes.NewReader(body))
					if err != nil {
						errs.Add(1)
						continue
					}
					var r struct{ ID string }
					b, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					_ = json.Unmarshal(b, &r)
					if resp.StatusCode >= 300 {
						fmt.Fprintln(os.Stderr, resp.StatusCode, string(b))
					} else {
						lastID.Store(r.ID)
					}
					rec(time.Since(t), resp.StatusCode)
					n.Add(1)
				}
			})
		}
		wg.Wait()
		el := time.Since(start)
		fmt.Printf("uploads=%d errs=%d codes=%v in %.1fs -> %.2f uploads/s, %.1f MB/s in; p50=%v p95=%v max=%v last=%v\n",
			n.Load(), errs.Load(), codes, el.Seconds(), float64(n.Load())/el.Seconds(),
			float64(n.Load())*float64(len(body))/1e6/el.Seconds(), pct(lat, .5), pct(lat, .95), pct(lat, 1), lastID.Load())
	case "read":
		id := os.Args[3]
		secs, _ := strconv.Atoi(os.Args[5])
		// what the UI does: snapshot, resource list, every 7th resource, a log
		resp, err := client.Get(base + "/api/v1/snapshots/" + id + "/resources")
		if err != nil {
			panic(err)
		}
		var list struct {
			Items []struct{ Kind, Namespace, Name string } `json:"items"`
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		_ = json.Unmarshal(raw, &list)
		urls := []string{"/api/v1/snapshots/" + id, "/api/v1/snapshots/" + id + "/resources"}
		for i, it := range list.Items {
			if i%7 == 0 {
				ns := it.Namespace
				if ns == "" {
					ns = "_cluster"
				}
				urls = append(urls, fmt.Sprintf("/api/v1/snapshots/%s/resources/%s/%s/%s", id, it.Kind, ns, it.Name))
			}
		}
		urls = append(urls, "/api/v1/snapshots/"+id+"/logs/"+os.Args[6])
		fmt.Printf("resource list bytes=%d items=%d, %d urls in rotation\n", len(raw), len(list.Items), len(urls))
		deadline := start.Add(time.Duration(secs) * time.Second)
		var bytesIn atomic.Int64
		for w := range conc {
			wg.Go(func() {
				for i := w; time.Now().Before(deadline); i++ {
					t := time.Now()
					resp, err := client.Get(base + urls[i%len(urls)])
					if err != nil {
						errs.Add(1)
						continue
					}
					c, _ := io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					bytesIn.Add(c)
					rec(time.Since(t), resp.StatusCode)
					n.Add(1)
				}
			})
		}
		wg.Wait()
		el := time.Since(start)
		fmt.Printf("reads=%d errs=%d codes=%v -> %.0f req/s, %.1f MB/s out; p50=%v p95=%v p99=%v\n",
			n.Load(), errs.Load(), codes, float64(n.Load())/el.Seconds(), float64(bytesIn.Load())/1e6/el.Seconds(), pct(lat, .5), pct(lat, .95), pct(lat, .99))
	}
}
