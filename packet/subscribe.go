package packet

import (
	"bytes"
	"errors"
	"fmt"
)

// Subscribe represents SUBSRIBE packet.
type Subscribe struct {
	PacketID ID
	Topics   []Topic
}

var _ Packet = (*Subscribe)(nil)

// Encode returns serialized Subscribe packet.
func (p *Subscribe) Encode() ([]byte, error) {
	var (
		header = &header{
			Type: TSubscribe,
			QoS:  QAtLeastOnce,
		}
		packetID = p.PacketID.bytes()
		topics   []byte
	)
	topics, err := encodeTopics(p.Topics)
	if err != nil {
		return nil, err
	}
	return encode(header, packetID, topics)
}

// Decode deserializes []byte as Subscribe packet.
func (p *Subscribe) Decode(b []byte) error {
	d, err := newDecoder(b, TSubscribe)
	if err != nil {
		return err
	}
	if d.header.QoS != QAtLeastOnce || d.header.Dup || d.header.Retain {
		return errors.New("invalid flags for Subscribe packet")
	}
	packetID, err := d.readPacketID()
	if err != nil {
		return err
	}
	topics, err := d.readTopics()
	if err != nil {
		return err
	}
	if err := d.finish(); err != nil {
		return err
	}
	*p = Subscribe{
		PacketID: packetID,
		Topics:   topics,
	}
	return nil
}

// AddTopic adds a topic to SUBSRIBE packet.
func (p *Subscribe) AddTopic(topic Topic) {
	p.Topics = append(p.Topics, topic)
}

// SubACK represents SUBACK packet.
type SubACK struct {
	PacketID ID
	Results  []SubscribeResult
}

var _ Packet = (*SubACK)(nil)

// Encode returns serialized SubACK packet.
func (p *SubACK) Encode() ([]byte, error) {
	// a vector of granted QoS levels.
	b := make([]byte, len(p.Results))
	for i, r := range p.Results {
		b[i] = byte(r)
	}
	return encode(&header{Type: TSubACK}, p.PacketID.bytes(), b)
}

// Decode deserializes []byte as SubACK packet.
func (p *SubACK) Decode(b []byte) error {
	d, err := newDecoder(b, TSubACK)
	if err != nil {
		return err
	}
	packetID, err := d.readPacketID()
	if err != nil {
		return err
	}
	results, err := d.readSubscribeResults()
	if err != nil {
		return err
	}
	if err := d.finish(); err != nil {
		return err
	}
	*p = SubACK{
		PacketID: packetID,
		Results:  results,
	}
	return nil
}

// AddResult adds a result of SUBSRIBE to the topic.
func (p *SubACK) AddResult(r SubscribeResult) {
	p.Results = append(p.Results, r)
}

// SubscribeResult represents result of subscribe to topic.
type SubscribeResult uint8

const (
	// SubscribeAtMostOnce is "Success - Maximum QoS 0"
	SubscribeAtMostOnce SubscribeResult = 0x00

	// SubscribeAtLeastOnce is "Success - Maximum QoS 1"
	SubscribeAtLeastOnce = 0x01

	// SubscribeExactOnce is "Success - Maximum QoS 2"
	SubscribeExactOnce = 0x02

	// SubscribeFailure is "Failure"
	SubscribeFailure = 0x80
)

// Unsubscribe represents UNSUBSCRIBE packet.
type Unsubscribe struct {
	PacketID ID
	Topics   []string
}

var _ Packet = (*Unsubscribe)(nil)

// Encode returns serialized Unsubscribe packet.
func (p *Unsubscribe) Encode() ([]byte, error) {
	var (
		header = &header{
			Type: TUnsubscribe,
			QoS:  QAtLeastOnce,
		}
		packetID = p.PacketID.bytes()
		topics   bytes.Buffer
	)
	for i, t := range p.Topics {
		b := encodeString(t)
		if b == nil {
			return nil, fmt.Errorf("too long topic name in #%d", i)
		}
		_, err := topics.Write(b)
		if err != nil {
			return nil, err
		}
	}
	return encode(header, packetID, topics.Bytes())
}

// Decode deserializes []byte as Unsubscribe packet.
func (p *Unsubscribe) Decode(b []byte) error {
	d, err := newDecoder(b, TUnsubscribe)
	if err != nil {
		return err
	}
	if d.header.QoS != QAtLeastOnce || d.header.Dup || d.header.Retain {
		return errors.New("invalid flags for Unsubscribe packet")
	}
	packetID, err := d.readPacketID()
	if err != nil {
		return err
	}
	topics, err := d.readStrings()
	if err != nil {
		return err
	}
	if err := d.finish(); err != nil {
		return err
	}
	*p = Unsubscribe{
		PacketID: packetID,
		Topics:   topics,
	}
	return nil
}

// UnsubACK represents UNSUBACK packet.
type UnsubACK struct {
	PacketID ID
}

var _ Packet = (*UnsubACK)(nil)

// Encode returns serialized UnsubACK packet.
func (p *UnsubACK) Encode() ([]byte, error) {
	return encode(&header{Type: TUnsubACK}, p.PacketID.bytes())
}

// Decode deserializes []byte as UnsubACK packet.
func (p *UnsubACK) Decode(b []byte) error {
	d, err := newDecoder(b, TUnsubACK)
	if err != nil {
		return err
	}
	packetID, err := d.readPacketID()
	if err != nil {
		return err
	}
	if err := d.finish(); err != nil {
		return err
	}
	*p = UnsubACK{
		PacketID: packetID,
	}
	return nil
}

// Topic represents topics to subscribe.
type Topic struct {
	Filter       string
	RequestedQoS QoS
}

func encodeTopics(topics []Topic) ([]byte, error) {
	buf := bytes.Buffer{}
	for i, t := range topics {
		n := encodeString(t.Filter)
		if n == nil {
			return nil, fmt.Errorf("too long topic name in #%d", i)
		}
		_, err := buf.Write(n)
		if err != nil {
			return nil, err
		}
		err = buf.WriteByte(byte(t.RequestedQoS & 0x03))
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
