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

// +build !nonetdev
// +build linux freebsd openbsd dragonfly darwin

package collector

import (
	"errors"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	netdevIgnoredDevices = kingpin.Flag("collector.netdev.device-blacklist", "Regexp of net devices to blacklist (mutually exclusive to device-whitelist).").String()
	netdevAcceptDevices  = kingpin.Flag("collector.netdev.device-whitelist", "Regexp of net devices to whitelist (mutually exclusive to device-blacklist).").String()
)

type netDevCollector struct {
	subsystem   string
	filter      netDevFilter
	metricDescs map[string]*prometheus.Desc
}

func init() {
	registerCollector("netdev", defaultEnabled, NewNetDevCollector)
}

// NewNetDevCollector returns a new Collector exposing network device stats.
func NewNetDevCollector() (Collector, error) {
	if *netdevIgnoredDevices != "" && *netdevAcceptDevices != "" {
		return nil, errors.New("device-blacklist & accept-devices are mutually exclusive")
	}

	return &netDevCollector{
		subsystem:   "network",
		filter:      newNetDevFilter(*netdevIgnoredDevices, *netdevAcceptDevices),
		metricDescs: map[string]*prometheus.Desc{},
	}, nil
}

func (c *netDevCollector) Update(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}
	metrics := make(chan netDevMetric, 64)

	wg.Add(1)
	go func() {
		for metric := range metrics {
			ch <- prometheus.MustNewConstMetric(c.getDesc(metric.key), prometheus.CounterValue, metric.value, metric.dev)
		}
		wg.Done()
	}()

	err := getNetDevStats(metrics, &c.filter)
	close(metrics)
	wg.Wait()

	if err != nil {
		return fmt.Errorf("couldn't get netstats: %s", err)
	}

	return nil
}

// getDesc builds and returns the description for a device metric
func (c *netDevCollector) getDesc(name string) *prometheus.Desc {
	desc, ok := c.metricDescs[name]
	if !ok {
		desc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, c.subsystem, name+"_total"),
			fmt.Sprintf("Network device statistic %s.", name),
			[]string{"device"},
			nil,
		)
		c.metricDescs[name] = desc
	}

	return desc
}
