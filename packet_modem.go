package main

// import (
// 	"errors"

// 	dsp "github.com/redwood/liquid-dsp/liquiddsp"
// )

// type PacketModem struct {
// 	q          dsp.Qpacketmodem
// 	payloadLen uint32
// 	frameLen   uint32
// }

// func NewPacketModem(payloadLen uint32, crcCheck dsp.CrcScheme, fec0 dsp.FecScheme, fec1 dsp.FecScheme, modScheme dsp.ModulationScheme) *PacketModem {
// 	q := dsp.QpacketmodemCreate()
// 	dsp.QpacketmodemConfigure(q, payloadLen, crcCheck, fec0, fec1, int32(modScheme))
// 	dsp.QpacketmodemPrint(q)
// 	return &PacketModem{
// 		q:          q,
// 		frameLen:   dsp.QpacketmodemGetFrameLen(q),
// 		payloadLen: payloadLen,
// 	}
// }

// func (pm *PacketModem) Close() error {
// 	dsp.QpacketmodemDestroy(pm.q)
// 	return nil
// }

// func (pm *PacketModem) Encode(bs []byte) ([]complex64, error) {
// 	frameTx := make([]complex64, pm.frameLen)
// 	dsp.QpacketmodemEncode(pm.q, string(bs), frameTx)
// 	return frameTx, nil
// }

// func (pm *PacketModem) Decode(c64s []complex64) ([]byte, error) {
// 	decoded := make([]byte, pm.payloadLen)
// 	if !dsp.QpacketmodemDecode(q, c64s, decoded) > 0 {
// 		return decoded, errors.New("payload failed CRC check")
// 	}
// 	return decoded, nil
// }
