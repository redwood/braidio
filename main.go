package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/multierr"
)

const (
	DEFAULT_SAMPLE_RATE = 10 * MHz
	DEFAULT_FREQ_HZ     = 3425 * MHz
	DEFAULT_TXVGA_GAIN  = 10
	DEFAULT_VGA_GAIN    = 20
	DEFAULT_LNA_GAIN    = 8
)

func main() {
	// q := NewPacketModem(payloadLen, crcCheck, fec0, fec1, modScheme)
	// defer q.Close()

	var radio *HackRF
	var modem *FlexFrameModem
	var err error

	radio, err = NewHackRF(DEFAULT_SAMPLE_RATE, DEFAULT_FREQ_HZ, DEFAULT_FREQ_HZ, DEFAULT_TXVGA_GAIN, DEFAULT_VGA_GAIN, DEFAULT_LNA_GAIN, func(bs []byte) {
		modem.DecodeFrames(iqSamplesToComplex64s(bs))
	})
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	modem = NewFlexFrameModem(
		func(c64s []complex64) {
			radio.Transmit(complex64sToIQSamples(c64s))
		},
		func(bs []byte) {
			fmt.Println("RECV:", string(bs))
		},
	)

	err = radio.Start()
	if err != nil {
		panic(err)
	}
	defer radio.Close()

	err = modem.Start()
	if err != nil {
		panic(err)
	}
	defer modem.Close()

	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGILL, syscall.SIGFPE, syscall.SIGSEGV, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		runtime.LockOSThread()
		select {
		case <-chSignal:
			err := multierr.Combine(
				modem.Close(),
				radio.Close(),
				os.Stdin.Close(),
			)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")

		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		modem.EncodeFrames([]byte(line))
	}
}

// var (
// 	payloadTx  = []byte(`{"foo": "bar", "baz": 123, "quux": false}`)
// 	payloadLen = uint32(len(payloadTx))
// )

// func tx() {
// 	q := NewPacketModem(payloadLen, crcCheck, fec0, fec1, modScheme)
// 	defer q.Close()

// 	must(hackrf.Init())
// 	defer hackrf.Exit()

// 	device := &hackrf.Device{}
// 	must(hackrf.Open(&device))
// 	defer hackrf.Close(device)

// 	must(hackrf.SetSampleRate(device, DEFAULT_SAMPLE_RATE))
// 	must(hackrf.SetTxvgaGain(device, DEFAULT_TXVGA_GAIN))

// 	frameBytes, err := q.Encode(payloadTx)
// 	if err != nil {
// 		panic(err)
// 	}

// 	t := transmit{r: bytes.NewReader(frameBytes)}
// 	must(hackrf.StartTx(device, t.txCallback, nil))
// 	defer hackrf.StopTx(device)

// 	must(hackrf.SetFreq(device, DEFAULT_FREQ_HZ))

// 	go func() {
// 		defer t.kill()
// 		for hackrf.IsStreaming(device) > 0 {
// 			time.Sleep(10 * time.Millisecond)
// 		}
// 	}()

// 	chSignal := make(chan os.Signal, 1)
// 	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGILL, syscall.SIGFPE, syscall.SIGSEGV, syscall.SIGTERM, syscall.SIGABRT)

// 	select {
// 	case <-chSignal:
// 		t.kill()
// 	case <-t.chStop:
// 	}
// }

// func rx() {
// 	// create and configure packet encoder/decoder object
// 	q := dsp.QpacketmodemCreate()
// 	defer dsp.QpacketmodemDestroy(q)
// 	dsp.QpacketmodemConfigure(q, payloadLen, crcCheck, fec0, fec1, int32(modScheme))
// 	dsp.QpacketmodemPrint(q)

// 	// get frame length
// 	frameLen := dsp.QpacketmodemGetFrameLen(q)

// 	must(hackrf.Init())
// 	defer hackrf.Exit()

// 	device := &hackrf.Device{}
// 	must(hackrf.Open(&device))
// 	defer hackrf.Close(device)

// 	must(hackrf.SetSampleRate(device, DEFAULT_SAMPLE_RATE))
// 	must(hackrf.SetVgaGain(device, DEFAULT_VGA_GAIN))
// 	must(hackrf.SetLnaGain(device, DEFAULT_LNA_GAIN))

// 	var buf bytes.Buffer
// 	r := receive{w: &buf}
// 	must(hackrf.StartRx(device, t.rxCallback, nil))
// 	defer hackrf.StopRx(device)

// 	must(hackrf.SetFreq(device, DEFAULT_FREQ_HZ))

// 	go func() {
// 		defer r.kill()
// 		for hackrf.IsStreaming(device) > 0 {
// 			time.Sleep(10 * time.Millisecond)
// 		}
// 	}()

// 	chSignal := make(chan os.Signal, 1)
// 	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGILL, syscall.SIGFPE, syscall.SIGSEGV, syscall.SIGTERM, syscall.SIGABRT)

// 	select {
// 	case <-chSignal:
// 		r.kill()
// 	case <-r.chStop:
// 	}

// 	fmt.Println("received buffer length:", buf.Len())
// 	c64s := iqSamplesToComplex64s(buf.Bytes())

// 	var (
// 		payloadLen = uint32(len(payloadTx))
// 		payloadRx  = make([]byte, payloadLen)
// 	)

// 	// decode frame
// 	crcPass := dsp.QpacketmodemDecode(q, c64s, payloadRx) > 0

// 	// count errors
// 	numBitErrors := dsp.CountBitErrorsArray(payloadTx, payloadRx, payloadLen)

// 	// print results
// 	if crcPass {
// 		fmt.Printf("payload PASS, errors: %d / %d\n", numBitErrors, 8*payloadLen)
// 	} else {
// 		fmt.Printf("payload FAIL, errors: %d / %d\n", numBitErrors, 8*payloadLen)
// 	}

// 	fmt.Println("output =", string(payloadRx))

// }
