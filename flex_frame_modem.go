package main

import (
	"fmt"

	dsp "github.com/redwood/liquid-dsp/liquiddsp"
	"go.uber.org/multierr"
)

type FlexFrameModem struct {
	writer *dsp.FlexFrameGen
	reader *dsp.FlexFrameSync

	onWroteFrameBytes func(bs []complex64)
	onReceivedFrame   func(bs []byte)

	chEncode chan []byte
	chDecode chan []complex64
	chStop   chan struct{}
	chDone   chan struct{}
}

var FLEX_FRAME_MODEM_HEADER = []byte("bradio is l33t")

func NewFlexFrameModem(onWroteFrameBytes func(iq []complex64), onReceivedFrame func(bs []byte)) *FlexFrameModem {
	m := &FlexFrameModem{
		onWroteFrameBytes: onWroteFrameBytes,
		onReceivedFrame:   onReceivedFrame,
		chEncode:          make(chan []byte, 1000),
		chDecode:          make(chan []complex64, 1000),
		chStop:            make(chan struct{}),
		chDone:            make(chan struct{}, 2),
	}

	var writerProps dsp.FlexFrameGenProps
	writerProps.InitDefault()
	writerProps.Check = uint32(dsp.LiquidCrcNone) // data validity check
	writerProps.Fec0 = uint32(dsp.LiquidFecNone)  // inner FEC scheme
	writerProps.Fec1 = uint32(dsp.LiquidFecNone)  // outer FEC scheme
	writerProps.ModScheme = uint32(dsp.LiquidModemQpsk)

	m.writer = dsp.NewFlexFrameGen(&writerProps)
	m.reader = dsp.NewFlexFrameSync(m.receiveCallback)

	return m
}

func (m *FlexFrameModem) Start() error {
	go func() {
		defer func() { m.chDone <- struct{}{} }()
		for {
			select {
			case <-m.chStop:
				return
			case bs := <-m.chEncode:
				m.encodeFrames(bs)
			}
		}
	}()

	go func() {
		defer func() { m.chDone <- struct{}{} }()
		for {
			select {
			case <-m.chStop:
				return
			case c64s := <-m.chDecode:
				m.decodeFrames(c64s)
			}
		}
	}()

	return nil
}

func (m *FlexFrameModem) Close() error {
	close(m.chStop)
	<-m.chDone
	<-m.chDone
	return multierr.Append(
		m.writer.Close(),
		m.reader.Close(),
	)
}

func (m *FlexFrameModem) EncodeFrames(bs []byte) {
	select {
	case m.chEncode <- bs:
	case <-m.chStop:
	}
}

func (m *FlexFrameModem) encodeFrames(bs []byte) {
	fmt.Println("[modem] encode frames:", string(bs))

	err := m.writer.Assemble(FLEX_FRAME_MODEM_HEADER, bs)
	if err != nil {
		fmt.Println("[modem] ERR: while assembling frame:", err)
		return
	} else if !m.writer.IsAssembled() {
		fmt.Println("[modem] ERR: could not assemble frame")
		return
	}

	var frameComplete bool
	for !frameComplete {
		var buf [256]complex64
		frameComplete = m.writer.WriteSamples(buf[:])
		fmt.Println("[modem] wrote frame samples")
		m.onWroteFrameBytes(buf[:])
	}
	fmt.Println("[modem] encoded frame")
}

func (m *FlexFrameModem) DecodeFrames(c64s []complex64) {
	select {
	case m.chDecode <- c64s:
	case <-m.chStop:
	}
}

func (m *FlexFrameModem) decodeFrames(c64s []complex64) {
	err := m.reader.Execute(c64s)
	if err != nil {
		fmt.Println("ERR: while decoding frames:", err)
	}
}

func (m *FlexFrameModem) receiveCallback(
	header []byte,
	headerValid bool,
	payload []byte,
	payloadValid bool,
	stats dsp.FrameSyncStats,
) bool {
	fmt.Println("******** callback invoked")

	stats.Print()
	fmt.Println("    payload             :   ", payload)
	fmt.Println("    header crc          :   ", headerValid)
	fmt.Println("    payload length      :   ", len(payload))
	fmt.Println("    payload crc         :   ", payloadValid)

	m.onReceivedFrame(payload)
	return false
}
