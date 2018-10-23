package modules

import (
	"fmt"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func tcpParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	tcp := pkt.Layer(layers.LayerTypeTCP).(*layers.TCP)

	if sniParser(ip, pkt, tcp) {
		return
	}
	if ntlmParser(ip, pkt, tcp) {
		return
	}
	if httpParser(ip, pkt, tcp) {
		return
	}
	if verbose {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"tcp",
			fmt.Sprintf("%s:%s", ip.SrcIP, vPort(tcp.SrcPort)),
			fmt.Sprintf("%s:%s", ip.DstIP, vPort(tcp.DstPort)),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s:%s > %s:%s %s",
			tui.Wrap(tui.BACKLIGHTBLUE+tui.FOREBLACK, "tcp"),
			vIP(ip.SrcIP),
			vPort(tcp.SrcPort),
			vIP(ip.DstIP),
			vPort(tcp.DstPort),
			tui.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}

func udpParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	udp := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP)

	if dnsParser(ip, pkt, udp) {
		return
	}
	if mdnsParser(ip, pkt, udp) {
		return
	}
	if krb5Parser(ip, pkt, udp) {
		return
	}
	if upnpParser(ip, pkt, udp) {
		return
	}
	if verbose {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"udp",
			fmt.Sprintf("%s:%s", ip.SrcIP, vPort(udp.SrcPort)),
			fmt.Sprintf("%s:%s", ip.DstIP, vPort(udp.DstPort)),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s:%s > %s:%s %s",
			tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, "udp"),
			vIP(ip.SrcIP),
			vPort(udp.SrcPort),
			vIP(ip.DstIP),
			vPort(udp.DstPort),
			tui.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}

func unkParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	if verbose {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			pkt.TransportLayer().LayerType().String(),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s > %s %s",
			tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, pkt.TransportLayer().LayerType().String()),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			tui.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}

func mainParser(pkt gopacket.Packet, verbose bool) bool {
	// simple networking sniffing mode?
	nlayer := pkt.NetworkLayer()
	if nlayer != nil {
		if nlayer.LayerType() != layers.LayerTypeIPv4 {
			log.Debug("Unexpected layer type %s, skipping packet.", nlayer.LayerType())
			log.Debug("%s", pkt.Dump())
			return false
		}

		ip := nlayer.(*layers.IPv4)

		tlayer := pkt.TransportLayer()
		if tlayer == nil {
			log.Debug("Missing transport layer skipping packet.")
			log.Debug("%s", pkt.Dump())
			return false
		}

		if tlayer.LayerType() == layers.LayerTypeTCP {
			tcpParser(ip, pkt, verbose)
		} else if tlayer.LayerType() == layers.LayerTypeUDP {
			udpParser(ip, pkt, verbose)
		} else {
			unkParser(ip, pkt, verbose)
		}
		return true
	}
	if ok, radiotap, dot11 := packets.Dot11Parse(pkt); ok {
		// are we sniffing in monitor mode?
		dot11Parser(radiotap, dot11, pkt, verbose)
		return true
	}
	return false
}
