package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mdlayher/raw"
	"math/rand"
	"net"
	"time"
)

// Option -
type Option struct {
	Type layers.DHCPOpt
	Data []byte
}

// Client -
type Client struct {
	Iface *net.Interface
	xid   uint32
	cnx   *raw.Conn
}

func (client *Client) newPacket(msgType layers.DHCPMsgType) *layers.DHCPv4 {
	packet := layers.DHCPv4{
		Operation:    layers.DHCPOpRequest,
		HardwareType: layers.LinkTypeEthernet,
		ClientHWAddr: client.Iface.HardwareAddr,
		Xid:          client.xid, // Transaction ID
	}

	packet.Options = append(packet.Options, layers.DHCPOption{
		Type:   layers.DHCPOptMessageType,
		Data:   []byte{byte(msgType)},
		Length: 1,
	})

	return &packet
}

func (client *Client) sendRequest(dhcp *layers.DHCPv4) error {
	eth := layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
		SrcMAC:       client.Iface.HardwareAddr,
		DstMAC:       layers.EthernetBroadcast,
	}
	ip := layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    []byte{0, 0, 0, 0},
		DstIP:    []byte{255, 255, 255, 255},
		Protocol: layers.IPProtocolUDP,
	}
	udp := layers.UDP{
		SrcPort: 68,
		DstPort: 67,
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := udp.SetNetworkLayerForChecksum(&ip); err != nil {
		return err
	}
	err := gopacket.SerializeLayers(buf, opts, &eth, &ip, &udp, dhcp)
	if err != nil {
		return err
	}

	_, err = client.cnx.WriteTo(buf.Bytes(), &raw.Addr{HardwareAddr: eth.DstMAC})
	return err
}

func (client *Client) parsePacket(data []byte) *layers.DHCPv4 {
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
	dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4)

	if dhcpLayer == nil {
		return nil
	}
	return dhcpLayer.(*layers.DHCPv4)
}

func (client *Client) readResponse(msgTypes ...layers.DHCPMsgType) (layers.DHCPMsgType, *net.IP, error) {
	client.cnx.SetReadDeadline(time.Now().Add(time.Second * 5))
	recvBuf := make([]byte, 1500)
	for {
		_, _, err := client.cnx.ReadFrom(recvBuf)
		if err != nil {
			return 0, nil, err
		}
		packet := client.parsePacket(recvBuf)
		if packet == nil {
			continue
		}

		var msgType layers.DHCPMsgType
		var resIP net.IP

		if packet.Xid == client.xid && packet.Operation == layers.DHCPOpReply {
			for _, option := range packet.Options {
				switch option.Type {
				case layers.DHCPOptMessageType:
					if option.Length == 1 {
						msgType = layers.DHCPMsgType(option.Data[0])
					}
				case layers.DHCPOptServerID:
					resIP = option.Data
				}
			}
			for _, t := range msgTypes {
				if t == msgType {
					return msgType, &resIP, nil
				}
			}
		}
	}
}

// DiscoverServer -
func (client *Client) DiscoverServer() (*net.IP, error) {
	cnx, err := raw.ListenPacket(client.Iface, uint16(layers.EthernetTypeIPv4), nil)
	if err != nil {
		return nil, err
	}
	client.cnx = cnx
	client.xid = rand.Uint32()

	defer func() {
		client.cnx.Close()
		client.cnx = nil
	}()

	err = client.sendRequest(client.newPacket(layers.DHCPMsgTypeDiscover))
	if err != nil {
		return nil, err
	}

	_, ip, err := client.readResponse(layers.DHCPMsgTypeOffer)
	if err != nil {
		return nil, err
	}
	return ip, nil
}
