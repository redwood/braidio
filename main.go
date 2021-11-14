package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var re = regexp.MustCompile(`(\/dev\/pts\/\d+)`)

func main() {
	cmdDirewolf := exec.Command("direwolf", strings.Split(`-t 0 -p -B 300`, " ")...)
	direwolfStdout, err := cmdDirewolf.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer direwolfStdout.Close()

	direwolfStderr, err := cmdDirewolf.StderrPipe()
	if err != nil {
		panic(err)
	}
	defer direwolfStderr.Close()

	go func() {
		scanner := bufio.NewScanner(direwolfStderr)
		for scanner.Scan() {
			matches := re.FindAllStringSubmatch(scanner.Text(), -1)
			for _, x := range matches {
				fmt.Println("=====")
				for _, y := range x {
					fmt.Println("-", y)
				}

			}

			fmt.Println("[direwolf]", scanner.Text())
		}
	}()

	go pipeOutputToTerminal("direwolf", direwolfStderr)

	err = cmdDirewolf.Start()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(3 * time.Second)
		transmit(direwolfStdout, fmt.Sprintf("howdy %v", time.Now()))
	}
}

func transmit(direwolfStdout io.ReadCloser, msg string) {
	cmdHackRF := exec.Command("hackrf_transfer", strings.Split(`-S 100 -p 1 -a 1 -t -`, " ")...)

	hackRFStdin, err := cmdHackRF.StdinPipe()
	if err != nil {
		panic(err)
	}

	hackRFStderr, err := cmdHackRF.StderrPipe()
	if err != nil {
		panic(err)
	}
	defer hackRFStderr.Close()

	go pipeOutputToTerminal("hackrf", hackRFStderr)

	go func() {
		_, err := io.Copy(hackRFStdin, direwolfStdout)
		if err != nil {
			panic(err)
		}
	}()

	err = cmdHackRF.Run()
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	bs := append([]byte{0xC0, 0x00}, []byte(msg)...)
	bs = append(bs, 0xC0)

	n, err := conn.Write(bs)
	if err != nil {
		panic(err)
	} else if n < 7 {
		panic("yeet")
	}
}

func pipeOutputToTerminal(label string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(fmt.Sprintf("[%v]"), scanner.Text())
	}
}
