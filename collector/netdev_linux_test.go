// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"io"
	"log"
	"os"
	"sync"
	"testing"
)

func TestNetDevStatsIgnore(t *testing.T) {
	file, err := os.Open("fixtures/proc/net/dev")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	filter := newNetDevFilter("^veth", "")

	netStats := collectNetDevStats(t, file, &filter)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(netStats)

	if want, got := float64(10437182923), netStats["wlan0/receive_bytes"]; want != got {
		t.Errorf("want netstat wlan0 bytes %v, got %v", want, got)
	}

	if want, got := float64(68210035552), netStats["eth0/receive_bytes"]; want != got {
		t.Errorf("want netstat eth0 bytes %v, got %v", want, got)
	}

	if want, got := float64(934), netStats["tun0/transmit_packets"]; want != got {
		t.Errorf("want netstat tun0 packets %v, got %v", want, got)
	}

	if want, got := 144, len(netStats); want != got {
		t.Errorf("want count of metrics to be %d, got %d", want, got)
	}

	if _, ok := netStats["veth4B09XN/transmit_bytes"]; ok {
		t.Error("want fixture interface veth4B09XN to not exist, but it does")
	}

	if want, got := float64(0), netStats["ibr10:30/receive_fifo"]; want != got {
		t.Error("want fixture interface ibr10:30 to exist, but it does not")
	}

	if want, got := float64(72), netStats["💩0/receive_multicast"]; want != got {
		t.Error("want fixture interface 💩0 to exist, but it does not")
	}
}

func collectNetDevStats(t *testing.T, reader io.Reader, f *netDevFilter) map[string]float64 {
	wg := sync.WaitGroup{}
	result := make(map[string]float64)
	c := make(chan netDevMetric)

	wg.Add(1)
	go func() {
		for metric := range c {
			result[metric.dev+"/"+metric.key] = metric.value
		}
		wg.Done()
	}()

	err := parseNetDevStats(c, reader, f)
	close(c)
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	return result
}
