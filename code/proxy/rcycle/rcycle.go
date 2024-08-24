package rcycle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/fxamacker/cbor"
	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
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

	CloseChan chan struct{}
}

func (rc *RequestCycle) ControlLoop(wg *sync.WaitGroup, isRequestType bool) {

	pp.Println("@ControlLoop/1")

	closables := make([]io.Closer, 0)

	defer func() {

		for _, closable := range closables {
			closable.Close()
		}

		wg.Done()
	}()

	for {

		pp.Println("@ControlLoop/2")

		select {
		case outreq := <-rc.OutsidePacket:
			pp.Println("@ControlLoop/gotOutsidePacket/3")

			rc.processPacket(*outreq.Packet, outreq.FromStream)
			if outreq.FromStream != nil {
				closables = append(closables, outreq.FromStream)
			}

		case <-rc.Context.Done():
			pp.Println("@ControlLoop/Context.Done/4")
			return
		case <-rc.CloseChan:
			pp.Println("@ControlLoop/CloseChan/5")
			return

		}

		if isRequestType {
			if rc.doneFragmentRecv() {
				pp.Println("@ControlLoop/doneFragmentRecv/6")
				break
			}
		}

		pp.Println("@ControlLoop/7")

	}

}

func (rc *RequestCycle) Close() {
	pp.Println("@Close/1")
	rc.CloseChan <- struct{}{}
	rc.ActiveStream.Close()
}

func (rc *RequestCycle) StreamWriteLoop() error {

	pp.Println("@StreamWriteLoop/1")

	fragmentId := uint32(0)
	counter := uint32(0)
	writeErrorCount := uint32(0)

	pp.Println("@StreamWriteLoop/2")

	for {
		counter++

		pp.Println("@StreamWriteLoop/3", counter)

		if fragmentId == rc.TotalFragments {
			pp.Println("@StreamWriteLoop/4")
			break
		}

		pp.Println("@StreamWriteLoop/5")

		if counter > rc.TotalFragments*3 {
			rc.ActiveStream.Close()
			return errors.New("timeout 1")
		}

		pp.Println("@StreamWriteLoop/6")

		// rotate stream if we have too many errors
		if writeErrorCount > 2 {
			rc.ResetStream()
			writeErrorCount = 0
		}

		pp.Println("@StreamWriteLoop/7")

		startOffset := fragmentId * wire.PacketFragmentationSize
		endOffset := fragmentId*wire.PacketFragmentationSize + wire.PacketFragmentationSize

		pp.Println("@StreamWriteLoop/8")

		if endOffset > uint32(len(rc.OutData)) {
			endOffset = uint32(len(rc.OutData))
		}

		pp.Println("@StreamWriteLoop/9")

		packet := &wire.Packet{
			PacketType:     wire.FragmentSend,
			HttpRequestId:  rc.RequestId,
			FragmentId:     fragmentId,
			TotalFragments: rc.TotalFragments,
			Data:           rc.OutData[startOffset:endOffset],
		}

		pp.Println("@StreamWriteLoop/10")

		pout, err := cbor.Marshal(packet, cbor.EncOptions{})
		if err != nil {
			pp.Println("@StreamWriteLoop/err_err", err.Error())
			panic(err)
		}

		pp.Println("@StreamWriteLoop/11")

		_, err = io.Copy(rc.ActiveStream, bytes.NewReader(pout))
		if err != nil {
			writeErrorCount++
			continue
		}

		pp.Println("@StreamWriteLoop/12")

		fragmentId++
	}

	pp.Println("@StreamWriteLoop/13")

	tallyPacket := &wire.Packet{
		PacketType:     wire.FragmentTally,
		HttpRequestId:  rc.RequestId,
		FragmentId:     0,
		TotalFragments: rc.TotalFragments,
		Data:           nil,
	}

	pp.Println("@StreamWriteLoop/14")

	tbyte, err := cbor.Marshal(tallyPacket, cbor.EncOptions{})
	if err != nil {
		panic(err)
	}

	pp.Println("@StreamWriteLoop/15")

	counter = uint32(0)
	writeErrorCount = uint32(0)

	for {
		counter++

		pp.Println("@StreamWriteLoop/16", counter)

		if counter > 5 {
			rc.ActiveStream.Close()
			pp.Println("@counter>5/timeout")
			return errors.New("timeout 2")
		}

		if writeErrorCount > 2 {
			pp.Println("@writeErrorCount>2")
			rc.ResetStream()
			writeErrorCount = 0
		}

		pp.Println("@StreamWriteLoop/17")

		_, err = io.Copy(rc.ActiveStream, bytes.NewBuffer(tbyte))
		if err != nil {
			writeErrorCount++
			pp.Println("@StreamWriteLoop/19/err", err.Error())
			continue
		}

		pp.Println("@StreamWriteLoop/18")

		break

	}

	return nil
}

func (rc *RequestCycle) ResetStream() {
	// FIXME => MUTEXT LOCK

	pp.Println("@ResetStream/1")

	rc.ActiveStream.Close()

	stream, err := rc.LocalNode.NewStream(rc.Context, rc.TargetPeer, mesh.ProtocolHttp2)
	if err != nil {
		panic(err)
	}

	pp.Println("@ResetStream/2")

	rc.ActiveStream = stream

}

func (rc *RequestCycle) StreamReadLoop(stream network.Stream) error {

	pp.Println("@StreamReadLoop/1")

	for {

		pp.Println("@StreamReadLoop/2")

		rPacket := wire.Packet{}

		m := cbor.NewDecoder(stream)

		pp.Println("@StreamReadLoop/3")

		err := m.Decode(&rPacket)
		if err != nil {
			pp.Println("@StreamReadLoop/err_err", err.Error())

			return err
		}

		pp.Println("@StreamReadLoop/4", rPacket.PacketType)

		rc.OutsidePacket <- SideChannelPacket{
			Packet:     &rPacket,
			FromStream: nil,
		}

		if rc.doneFragmentRecv() {
			pp.Println("@StreamReadLoop/doneFragmentRecv/6")
			return nil
		}

		pp.Println("@StreamReadLoop/5")

	}

}

// utils

func (rc *RequestCycle) doneFragmentRecv() bool {

	pp.Println("@doneFragmentRecv/1")

	for idx, done := range rc.DonePacketFrags {
		if !done {
			return false
		}

		pp.Println("@doneFragmentRecv/2", idx, done)

	}

	return true
}

func (rc *RequestCycle) processPacket(rPacket wire.Packet, stream network.Stream) {

	pp.Println("@processPacket/1")

	writeErrorCount := uint32(0)

	if stream == nil {
		stream = rc.ActiveStream
	}

	pp.Println("@processPacket/2", rPacket.PacketType, stream == nil)

	if rPacket.PacketType == wire.FragmentResend {

		pp.Println("@processPacket/3")

		ids := make([]uint32, 0)
		for _, id := range strings.Split(string(rPacket.Data), ",") {
			id, _ := strconv.ParseUint(id, 10, 32)
			ids = append(ids, uint32(id))
		}

		pp.Println("@processPacket/4")

		for _, id := range ids {

			pp.Println("@processPacket/5", id)

			startOffset := id * wire.PacketFragmentationSize
			endOffset := id*wire.PacketFragmentationSize + wire.PacketFragmentationSize

			if endOffset > uint32(len(rc.OutData)) {
				endOffset = uint32(len(rc.OutData))
			}

			pp.Println("@processPacket/6")

			packet := &wire.Packet{
				PacketType:     wire.FragmentSend,
				HttpRequestId:  rc.RequestId,
				FragmentId:     id,
				TotalFragments: rc.TotalFragments,
				Data:           rc.OutData[startOffset:endOffset],
			}

			pp.Println("@processPacket/7")

			pout, err := cbor.Marshal(packet, cbor.EncOptions{})
			if err != nil {
				panic(err)
			}

			pp.Println("@processPacket/8")

			_, err = io.Copy(stream, bytes.NewBuffer(pout))
			if err != nil {
				writeErrorCount++
				continue
			}

			pp.Println("@processPacket/9")

		}

		pp.Println("@processPacket/10/return")

		return
	}

	if rPacket.PacketType == wire.FragmentSend {

		pp.Println("@processPacket/11")

		if rc.InData == nil {
			rc.InData = make([]byte, rPacket.TotalFragments*wire.PacketFragmentationSize)
			for i := 0; i < int(rPacket.TotalFragments); i++ {
				rc.DonePacketFrags[uint32(i)] = false
			}

			pp.Println("@processPacket/12")
		}

		pp.Println("@processPacket/13")

		startOffset := rPacket.FragmentId * wire.PacketFragmentationSize
		endOffset := rPacket.FragmentId*wire.PacketFragmentationSize + wire.PacketFragmentationSize

		pp.Println("@processPacket/14", startOffset, endOffset)

		if endOffset > uint32(len(rc.InData)) {
			endOffset = uint32(len(rc.InData))
		}

		pp.Println("@processPacket/15", startOffset, endOffset)

		copy(rc.InData[startOffset:endOffset], rPacket.Data)
		rc.DonePacketFrags[rPacket.FragmentId] = true

		pp.Println("@processPacket/16")

		return

	}

	if rPacket.PacketType == wire.FragmentTally {
		//check if all fragments are done
		pendingIds := strings.Builder{}
		isFirst := true

		pp.Println("@processPacket/17")

		for fid, done := range rc.DonePacketFrags {
			pp.Println("@processPacket/18", fid, done, isFirst)

			if !done {

				if isFirst {
					isFirst = false
				} else {
					pendingIds.WriteString(",")
				}

				pendingIds.WriteString(strconv.FormatUint(uint64(fid), 10))

			}

		}

		pp.Println("@processPacket/19")

		if len(pendingIds.String()) != 0 {

			pp.Println("@processPacket/20")

			packet := &wire.Packet{
				PacketType:     wire.FragmentResend,
				HttpRequestId:  rc.RequestId,
				FragmentId:     0,
				TotalFragments: rc.TotalFragments,
				Data:           []byte(pendingIds.String()),
			}

			pout, err := cbor.Marshal(packet, cbor.EncOptions{})
			if err != nil {
				panic(err)
			}

			io.Copy(stream, bytes.NewBuffer(pout))
		} else {

			pp.Println("@processPacket/21")

			paket := &wire.Packet{
				PacketType:     wire.FragmentEnd,
				HttpRequestId:  rc.RequestId,
				FragmentId:     0,
				TotalFragments: rc.TotalFragments,
				Data:           nil,
			}

			pout, err := cbor.Marshal(paket, cbor.EncOptions{})
			if err != nil {
				panic(err)
			}

			io.Copy(stream, bytes.NewBuffer(pout))

		}

		pp.Println("@processPacket/22/return")

		return

	}

	if rPacket.PacketType == wire.FragmentEnd {
		rc.CloseChan <- struct{}{}

		pp.Println("@processPacket/23/FragmentEnd")

		return
	}

}
