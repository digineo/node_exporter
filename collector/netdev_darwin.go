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
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/prometheus/common/log"
	"golang.org/x/sys/unix"
)

func getNetDevStats(stats netDevStats, f *netDevFilter) error {
	ifs, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("net.Interfaces() failed: %w", err)
	}

	for _, iface := range ifs {
		if f.ignored(iface.Name) {
			log.Debugf("Ignoring device: %s", iface.Name)
			continue
		}

		ifaceData, err := getIfaceData(iface.Index)
		if err != nil {
			log.Debugf("failed to load data for interface %q: %v", iface.Name, err)
			continue
		}

		stats.add(iface.Name, "receive_packets", float64(ifaceData.Data.Ipackets))
		stats.add(iface.Name, "transmit_packets", float64(ifaceData.Data.Opackets))
		stats.add(iface.Name, "receive_errs", float64(ifaceData.Data.Ierrors))
		stats.add(iface.Name, "transmit_errs", float64(ifaceData.Data.Oerrors))
		stats.add(iface.Name, "receive_bytes", float64(ifaceData.Data.Ibytes))
		stats.add(iface.Name, "transmit_bytes", float64(ifaceData.Data.Obytes))
		stats.add(iface.Name, "receive_multicast", float64(ifaceData.Data.Imcasts))
		stats.add(iface.Name, "transmit_multicast", float64(ifaceData.Data.Omcasts))
	}

	return nil
}

func getIfaceData(index int) (*ifMsghdr2, error) {
	var data ifMsghdr2
	rawData, err := unix.SysctlRaw("net", unix.AF_ROUTE, 0, 0, unix.NET_RT_IFLIST2, index)
	if err != nil {
		return nil, err
	}
	err = binary.Read(bytes.NewReader(rawData), binary.LittleEndian, &data)
	return &data, err
}

type ifMsghdr2 struct {
	Msglen    uint16
	Version   uint8
	Type      uint8
	Addrs     int32
	Flags     int32
	Index     uint16
	_         [2]byte
	SndLen    int32
	SndMaxlen int32
	SndDrops  int32
	Timer     int32
	Data      ifData64
}

type ifData64 struct {
	Type       uint8
	Typelen    uint8
	Physical   uint8
	Addrlen    uint8
	Hdrlen     uint8
	Recvquota  uint8
	Xmitquota  uint8
	Unused1    uint8
	Mtu        uint32
	Metric     uint32
	Baudrate   uint64
	Ipackets   uint64
	Ierrors    uint64
	Opackets   uint64
	Oerrors    uint64
	Collisions uint64
	Ibytes     uint64
	Obytes     uint64
	Imcasts    uint64
	Omcasts    uint64
	Iqdrops    uint64
	Noproto    uint64
	Recvtiming uint32
	Xmittiming uint32
	Lastchange unix.Timeval32
}
