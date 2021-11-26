package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/redwood/hackrf/hackrf"
	"go.uber.org/multierr"
)

type HackRF struct {
	device       *hackrf.Device
	sampleRateHz float64

	txFreqHz  uint
	txvgaGain uint32

	rxFreqHz uint
	vgaGain  uint32
	lnaGain  uint32

	chTx chan []byte
	onRx func(bs []byte)

	stopOnce sync.Once
	chStop   chan struct{}
	chDone   chan struct{}
}

func NewHackRF(sampleRateHz float64, txFreqHz, rxFreqHz uint, txvgaGain, vgaGain, lnaGain uint32, onRx func(bs []byte)) (*HackRF, error) {
	err := hackrf.Init()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			hackrf.Exit()
		}
	}()

	device, err := hackrf.NewDevice()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			device.Close()
		}
	}()

	err = device.SetSampleRate(sampleRateHz)
	if err != nil {
		return nil, errors.Wrap(err, "while setting sample rate")
	}
	err = device.SetTxvgaGain(txvgaGain)
	if err != nil {
		return nil, errors.Wrap(err, "while setting txvga gain")
	}
	err = device.SetVgaGain(vgaGain)
	if err != nil {
		return nil, errors.Wrap(err, "while setting vga gain")
	}
	err = device.SetLnaGain(lnaGain)
	if err != nil {
		return nil, errors.Wrap(err, "while setting lna gain")
	}
	return &HackRF{
		device:       device,
		sampleRateHz: sampleRateHz,
		txFreqHz:     txFreqHz,
		txvgaGain:    txvgaGain,
		rxFreqHz:     rxFreqHz,
		vgaGain:      vgaGain,
		lnaGain:      lnaGain,
		chTx:         make(chan []byte),
		onRx:         onRx,
		chStop:       make(chan struct{}),
		chDone:       make(chan struct{}),
	}, nil
}

func (r *HackRF) Start() error {
	go r.runloop()
	return nil
}

func (r *HackRF) Close() error {
	var err error
	r.stopOnce.Do(func() {
		close(r.chStop)
		<-r.chDone
		err = multierr.Combine(
			r.device.StopRx(),
			r.device.StopTx(),
			r.device.Close(),
			hackrf.Exit(),
		)
	})
	return err
}

func (r *HackRF) runloop() {
	defer close(r.chDone)

	for {
		if !r.startRx() {
			return
		}

		select {
		case msg := <-r.chTx:
			r.transmit(msg)
		case <-r.chStop:
			return
		}
	}
}

func (r *HackRF) Transmit(bs []byte) {
	select {
	case r.chTx <- bs:
	case <-r.chStop:
	}
}

func (r *HackRF) startRx() (ok bool) {
	fmt.Println("[radio] start rx")

	err := r.device.StartRx(r.rxCallback)
	if err != nil {
		fmt.Println("[radio] ERR: while starting rx:", err)
		return false
	}

	err = r.device.SetFreq(r.rxFreqHz)
	if err != nil {
		fmt.Println("[radio] ERR: while setting frequency", err)
		r.stopRx()
		return false
	}

	for !r.device.IsStreaming() {
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("[radio] start rx DONE")
	return true
}

func (r *HackRF) stopRx() (ok bool) {
	fmt.Println("[radio] stop rx")

	err := r.device.StopRx()
	if err != nil {
		fmt.Println("[radio] ERR: while stopping rx:", err)
		return false
	}

	for r.device.IsStreaming() {
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("[radio] stop rx DONE")
	return true
}

func (r *HackRF) startTx(callback hackrf.TransferCallback) (ok bool) {
	fmt.Println("[radio] start tx / is streaming =", r.device.IsStreaming())

	for {
		err := r.device.StartTx(callback)
		if err == nil {
			break
		}
		// if err != nil {
		// 	fmt.Println("[radio] ERR: while starting tx:", err)
		// 	return false
		// }
	}

	err := r.device.SetFreq(r.txFreqHz)
	if err != nil {
		fmt.Println("[radio] ERR: while setting frequency:", err)
		r.stopTx()
		return false
	}

	for !r.device.IsStreaming() {
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("[radio] start tx DONE")
	return true
}

func (r *HackRF) stopTx() (ok bool) {
	fmt.Println("[radio] stop tx")

	err := r.device.StopTx()
	if err != nil {
		fmt.Println("[radio] ERR: while stopping tx:", err)
		return false
	}

	for r.device.IsStreaming() {
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("[radio] stop tx DONE")
	return true
}

func (r *HackRF) transmit(bs []byte) {
	if !r.stopRx() {
		return
	}

	t := transmit{
		bs:     bs,
		chDone: make(chan struct{}),
	}

	if !r.startTx(t.txCallback) {
		t.kill()
		return
	}

	select {
	case <-t.chDone:
		fmt.Println("[radio] transmission complete")
	case <-r.chStop:
	}

	t.kill()
	r.stopTx()
}

type transmit struct {
	bs       []byte
	i        int
	chDone   chan struct{}
	stopOnce sync.Once
}

func (t *transmit) txCallback(transfer *hackrf.Transfer, err error) int32 {
	if err != nil {
		fmt.Println("[radio] ERR: in tx callback:", err)
		t.kill()
		return -1
	}

	select {
	case <-t.chDone:
		fmt.Println("[radio] tx callback done")
		return -1
	default:
	}

	fmt.Println("[radio] tx callback")

	if t.i >= len(t.bs) {
		fmt.Println("[radio] tx: no more to copy")
		// Done
		t.kill()
		return -1
	}

	end := t.i + int(transfer.ValidLength)
	if end > len(t.bs) {
		end = len(t.bs)
	}

	n := copy(transfer.Buffer[:transfer.ValidLength], t.bs[t.i:end])
	t.i += n
	if n < int(transfer.ValidLength) {
		fmt.Println("[radio] tx: copying final bytes")
		t.kill()
		return 0
	}
	fmt.Println("[radio] tx: copied", n, "bytes")
	return 0
}

func (t *transmit) kill() {
	t.stopOnce.Do(func() { close(t.chDone) })
}

func (r *HackRF) rxCallback(transfer *hackrf.Transfer, err error) int32 {
	if err != nil {
		fmt.Println("[radio] ERR: in rx callback:", err)
		return -1
	}

	select {
	case <-r.chStop:
		return -1
	default:
	}

	bs := make([]byte, transfer.ValidLength)
	n := copy(bs, transfer.Buffer[:transfer.ValidLength])

	r.onRx(bs)
	fmt.Println("[radio] rx: copied", n, "bytes")
	return 0
}
