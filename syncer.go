package main

// import "bytes"

// type Syncer struct {
// 	startWord [8]byte
// 	stopWord  [8]byte
// }

// func (s Syncer) scanPackets(data []byte, atEOF bool) (advance int, token []byte, err error) {
// 	if atEOF && len(data) == 0 {
// 		return 0, nil, nil
// 	}

// 	if start := bytes.Index(data, s.startWord); start >= 0 {
// 		if stop := bytes.Index(data, s.stopWord); stop > start {
// 			return stop + len(s.stopWord) + 1, data[start+len(s.startWord) : stop], nil
// 		}
// 	}
// 	// Request more data.
// 	return 0, nil, nil
// }
