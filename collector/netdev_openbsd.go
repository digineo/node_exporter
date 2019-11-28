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

package collector

import (
	"errors"
	"regexp"

	"github.com/prometheus/common/log"
)

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <ifaddrs.h>
#include <net/if.h>
*/
import "C"

func getNetDevStats(ignore *regexp.Regexp, accept *regexp.Regexp) (netDevStats, error) {
	netDev := netDevStats{}

	var ifap, ifa *C.struct_ifaddrs
	if C.getifaddrs(&ifap) == -1 {
		return nil, errors.New("getifaddrs() failed")
	}
	defer C.freeifaddrs(ifap)

	for ifa = ifap; ifa != nil; ifa = ifa.ifa_next {
		if ifa.ifa_addr.sa_family == C.AF_LINK {
			dev := C.GoString(ifa.ifa_name)
			if ignore != nil && ignore.MatchString(dev) {
				log.Debugf("Ignoring device: %s", dev)
				continue
			}
			if accept != nil && !accept.MatchString(dev) {
				log.Debugf("Ignoring device: %s", dev)
				continue
			}

			data := (*C.struct_if_data)(ifa.ifa_data)

			netDev[dev] = map[string]uint64{
				"receive_packets":    uint64(data.ifi_ipackets),
				"transmit_packets":   uint64(data.ifi_opackets),
				"receive_errs":       uint64(data.ifi_ierrors),
				"transmit_errs":      uint64(data.ifi_oerrors),
				"receive_bytes":      uint64(data.ifi_ibytes),
				"transmit_bytes":     uint64(data.ifi_obytes),
				"receive_multicast":  uint64(data.ifi_imcasts),
				"transmit_multicast": uint64(data.ifi_omcasts),
				"receive_drop":       uint64(data.ifi_iqdrops),
			}
		}
	}

	return netDev, nil
}
