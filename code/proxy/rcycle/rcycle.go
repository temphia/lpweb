package rcycle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/polydawn/refmt/cbor"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/wire"
)

type SideChannelPacket struct {
	Packet     *wire.Packet
	FromStream network.Stream
}

type RequestCycle struct {
	Context   context.Context
	RequestId uint32
	LocalNode host.Host

	OutsidePacket  chan SideChannelPacket
	OutData        []byte
	ActiveStream   network.Stream
	TotalFragments uint32
	TargetPeer     peer.ID

	DonePacketFrags map[uint32]bool
	DoneInDataChan  chan uint32
	InData          []byte
}

func (rc *RequestCycle) ControlLoop() {

	for {
		select {
		case outreq := <-rc.OutsidePacket:
			rc.processPacket(*outreq.Packet, outreq.FromStream)
		default:

		}

		if rc.doneFragmentRecv() {
			break
		}

	}

}

func (rc *RequestCycle) StreamWriteLoop() error {

	fragmentId := uint32(0)
	counter := uint32(0)
	writeErrorCount := uint32(0)

	for {
		counter++

		if fragmentId == rc.TotalFragments {
			break
		}

		if counter > rc.TotalFragments*3 {
			rc.ActiveStream.Close()
			return errors.New("timeout 1")
		}

		// rotate stream if we have too many errors
		if writeErrorCount > 2 {
			rc.ResetStream()
			writeErrorCount = 0
		}

		startOffset := fragmentId * wire.PacketFragmentationSize
		endOffset := fragmentId*wire.PacketFragmentationSize + wire.PacketFragmentationSize

		if endOffset > uint32(len(rc.OutData)) {
			endOffset = uint32(len(rc.OutData))
		}

		packet := wire.Packet{
			PacketType:     wire.FragmentSend,
			HttpRequestId:  rc.RequestId,
			FragmentId:     fragmentId,
			TotalFragments: rc.TotalFragments,
			Data:           rc.OutData[startOffset:endOffset],
		}

		pout, err := cbor.Marshal(packet)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(rc.ActiveStream, bytes.NewBuffer(pout))
		if err != nil {
			writeErrorCount++
			continue
		}

		fragmentId++
	}

	tallyPacket := wire.Packet{
		PacketType:     wire.FragmentSend,
		HttpRequestId:  rc.RequestId,
		FragmentId:     0,
		TotalFragments: rc.TotalFragments,
		Data:           nil,
	}

	tbyte, err := cbor.Marshal(tallyPacket)
	if err != nil {
		panic(err)
	}

	counter = uint32(0)
	writeErrorCount = uint32(0)

	for {
		counter++

		if counter > 5 {
			rc.ActiveStream.Close()
			return errors.New("timeout 2")
		}

		if writeErrorCount > 2 {
			rc.ResetStream()
			writeErrorCount = 0
		}

		_, err = io.Copy(rc.ActiveStream, bytes.NewBuffer(tbyte))
		if err != nil {
			writeErrorCount++
			continue
		}

		break

	}

	return nil
}

func (rc *RequestCycle) ResetStream() {
	// FIXME => MUTEXT LOCK

	rc.ActiveStream.Close()

	stream, err := rc.LocalNode.NewStream(rc.Context, rc.TargetPeer, mesh.ProtocolHttp2)
	if err != nil {
		panic(err)
	}

	rc.ActiveStream = stream

}

func (rc *RequestCycle) StreamReadLoop(stream network.Stream) error {

	for {

		rPacket := wire.Packet{}

		m := cbor.NewUnmarshaller(cbor.DecodeOptions{
			CoerceUndefToNull: true,
		}, stream)

		err := m.Unmarshal(&rPacket)
		if err != nil {
			return err
		}

		rc.OutsidePacket <- SideChannelPacket{
			Packet:     &rPacket,
			FromStream: nil,
		}

	}

}

// utils

func (rc *RequestCycle) doneFragmentRecv() bool {

	for _, done := range rc.DonePacketFrags {
		if !done {
			return false
		}

	}

	return true
}

func (rc *RequestCycle) processPacket(rPacket wire.Packet, stream network.Stream) {

	writeErrorCount := uint32(0)

	if stream == nil {
		stream = rc.ActiveStream
	}

	if rPacket.PacketType == wire.FragmentResend {
		ids := make([]uint32, 0)
		for _, id := range strings.Split(string(rPacket.Data), ",") {
			id, _ := strconv.ParseUint(id, 10, 32)
			ids = append(ids, uint32(id))
		}

		for _, id := range ids {

			startOffset := id * wire.PacketFragmentationSize
			endOffset := id*wire.PacketFragmentationSize + wire.PacketFragmentationSize

			if endOffset > uint32(len(rc.OutData)) {
				endOffset = uint32(len(rc.OutData))
			}

			packet := &wire.Packet{
				PacketType:     wire.FragmentSend,
				HttpRequestId:  rc.RequestId,
				FragmentId:     id,
				TotalFragments: rc.TotalFragments,
				Data:           rc.OutData[startOffset:endOffset],
			}

			pout, err := cbor.Marshal(packet)
			if err != nil {
				panic(err)
			}

			_, err = io.Copy(stream, bytes.NewBuffer(pout))
			if err != nil {
				writeErrorCount++
				continue
			}

		}

		return
	}

	if rPacket.PacketType == wire.FragmentSend {

		if rc.InData == nil {
			rc.InData = make([]byte, rPacket.TotalFragments*wire.PacketFragmentationSize)
			for i := 0; i < int(rPacket.TotalFragments); i++ {
				rc.DonePacketFrags[uint32(i)] = false
			}
		}

		startOffset := rPacket.FragmentId * wire.PacketFragmentationSize
		endOffset := rPacket.FragmentId*wire.PacketFragmentationSize + wire.PacketFragmentationSize

		if endOffset > uint32(len(rc.InData)) {
			endOffset = uint32(len(rc.InData))
		}

		copy(rc.InData[startOffset:endOffset], rPacket.Data)
		rc.DonePacketFrags[rPacket.FragmentId] = true

		return

	}

	if rPacket.PacketType == wire.FragmentTally {
		//check if all fragments are done
		pendingIds := strings.Builder{}
		isFirst := true

		for fid, done := range rc.DonePacketFrags {
			if !done {

				if isFirst {
					isFirst = false
				} else {
					pendingIds.WriteString(",")
				}

				pendingIds.WriteString(strconv.FormatUint(uint64(fid), 10))

			}

		}

		packet := &wire.Packet{
			PacketType:     wire.FragmentResend,
			HttpRequestId:  rc.RequestId,
			FragmentId:     0,
			TotalFragments: rc.TotalFragments,
			Data:           []byte(pendingIds.String()),
		}

		pout, err := cbor.Marshal(packet)
		if err != nil {
			panic(err)
		}

		io.Copy(stream, bytes.NewBuffer(pout))
	}

}
