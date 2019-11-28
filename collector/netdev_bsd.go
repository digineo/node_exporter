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
// +build freebsd dragonfly

package collector

import (
	"errors"

	"github.com/prometheus/common/log"
)

/*
#cgo CFLAGS: -D_IFI_OQDROPS
#include <stdio.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <ifaddrs.h>
#include <net/if.h>
*/
import "C"

func getNetDevStats(stats netDevStats, f *netDevFilter) error {
	var ifap, ifa *C.struct_ifaddrs
	if C.getifaddrs(&ifap) == -1 {
		return errors.New("getifaddrs() failed")
	}
	defer C.freeifaddrs(ifap)

	for ifa = ifap; ifa != nil; ifa = ifa.ifa_next {
		if ifa.ifa_addr.sa_family != C.AF_LINK {
			continue
		}

		dev := C.GoString(ifa.ifa_name)
		if f.ignored(dev) {
			log.Debugf("Ignoring device: %s", dev)
			continue
		}

		data := (*C.struct_if_data)(ifa.ifa_data)

		stats.add(dev, "receive_packets", float64(data.ifi_ipackets))
		stats.add(dev, "transmit_packets", float64(data.ifi_opackets))
		stats.add(dev, "receive_errs", float64(data.ifi_ierrors))
		stats.add(dev, "transmit_errs", float64(data.ifi_oerrors))
		stats.add(dev, "receive_bytes", float64(data.ifi_ibytes))
		stats.add(dev, "transmit_bytes", float64(data.ifi_obytes))
		stats.add(dev, "receive_multicast", float64(data.ifi_imcasts))
		stats.add(dev, "transmit_multicast", float64(data.ifi_omcasts))
		stats.add(dev, "receive_drop", float64(data.ifi_iqdrops))
		stats.add(dev, "transmit_drop", float64(data.ifi_oqdrops))
	}

	return nil
}
