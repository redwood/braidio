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
	var direwolfTxStdout io.ReadCloser
	var direwolfTxStdin io.WriteCloser
	{
		var err error
		cmdDirewolfTx := exec.Command("/usr/local/bin/direwolf", strings.Split(`-t 0 -p -B 300 -c /home/spook/direwolf-tx.conf`, " ")...)
		direwolfTxStdout, err = cmdDirewolfTx.StdoutPipe()
		if err != nil {
			panic(err)
		}
		defer direwolfTxStdout.Close()

		direwolfTxStdin, err = cmdDirewolfTx.StdinPipe()
		if err != nil {
			panic(err)
		}
		defer direwolfTxStdin.Close()

		direwolfStderr, err := cmdDirewolfTx.StderrPipe()
		if err != nil {
			panic(err)
		}
		defer direwolfStderr.Close()
		go pipeOutputToTerminal("direwolf-tx", direwolfStderr)

		err = cmdDirewolfTx.Start()
		if err != nil {
			panic(err)
		}
	}
	defer direwolfTxStdin.Close()

	var direwolfRxStdout io.ReadCloser
	var direwolfRxStdin io.WriteCloser
	{
		var err error
		cmdDirewolfRx := exec.Command("/usr/local/bin/direwolf", strings.Split(`-t 0 -p -B 300 -c /home/spook/direwolf-rx.conf`, " ")...)
		direwolfRxStdout, err = cmdDirewolfRx.StdoutPipe()
		if err != nil {
			panic(err)
		}
		defer direwolfRxStdout.Close()

		direwolfRxStdin, err = cmdDirewolfRx.StdinPipe()
		if err != nil {
			panic(err)
		}
		defer direwolfRxStdin.Close()

		direwolfStderr, err := cmdDirewolfRx.StderrPipe()
		if err != nil {
			panic(err)
		}
		defer direwolfStderr.Close()
		go pipeOutputToTerminal("direwolf-rx", direwolfStderr)

		err = cmdDirewolfRx.Start()
		if err != nil {
			panic(err)
		}
	}
	defer direwolfRxStdin.Close()

	time.Sleep(1 * time.Second)

	connTx, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		panic(err)
	}
	defer connTx.Close()

	connRx, err := net.Dial("tcp", "127.0.0.1:9001")
	if err != nil {
		panic(err)
	}
	defer connRx.Close()

	time.Sleep(1 * time.Second)

	go func() {
		for {
			transmit(direwolfTxStdout, connTx, fmt.Sprintf("howdy %v", time.Now()))
			time.Sleep(3 * time.Second)
		}
	}()

	// go receive(direwolfTxStdin, direwolfTxStdout, conn)
	{

	}

	select {}
}

func direwolfTx() (stdin io.WriteCloser, stdout io.ReadCloser) {
	var err error
	cmd := exec.Command("/usr/local/bin/direwolf", strings.Split(`-t 0 -p -B 300 -c /home/spook/direwolf-tx.conf`, " ")...)
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdout.Close()

	stdin, err = cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer stdin.Close()

	direwolfStderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	defer direwolfStderr.Close()
	go pipeOutputToTerminal("direwolf-tx", direwolfStderr)

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	return stdin, stdout
}

func direwolfRx() (stdin io.WriteCloser, stdout io.ReadCloser) {
	var err error
	cmd := exec.Command("/usr/local/bin/direwolf", strings.Split(`-t 0 -p -B 300 -c /home/spook/direwolf-rx.conf`, " ")...)
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdout.Close()

	stdin, err = cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer stdin.Close()

	direwolfStderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	defer direwolfStderr.Close()
	go pipeOutputToTerminal("direwolf-rx", direwolfStderr)

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	return stdin, stdout
}

func transmit(direwolfTxStdout io.Reader, conn net.Conn, msg string) {
	cmdHackRF := exec.Command("hackrf_transfer", strings.Split(`-S 100 -p 1 -a 1 -t -`, " ")...)
	go pipeStderrToTerminal("hackrf", cmdHackRF)

	hackRFStdin, err := cmdHackRF.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer hackRFStdin.Close()

	go func() {
		_, err := io.Copy(hackRFStdin, direwolfTxStdout)
		if err != nil {
			time.Sleep(1 * time.Second)
			panic(err)
		}
	}()

	err = cmdHackRF.Run()
	if err != nil {
		panic(err)
	}

	bs := append([]byte{0xC0, 0x00}, []byte(msg)...)
	bs = append(bs, 0xC0)

	n, err := conn.Write(bs)
	if err != nil {
		panic(err)
	} else if n < 7 {
		panic("yeet")
	}
}

func receive(direwolfRxStdin io.Writer, conn net.Conn) {
	cmdHackRF := exec.Command("hackrf_transfer", strings.Split(`-S 1200 -p 1 -a 1 -r -`, " ")...)
	go pipeStderrToTerminal("hackrf", cmdHackRF)

	hackRFStdout, err := cmdHackRF.StdoutPipe()
	if err != nil {
		panic(err)
	}

	go func() {
		_, err := io.Copy(direwolfRxStdin, hackRFStdout)
		if err != nil {
			panic(err)
		}
	}()

	err = cmdHackRF.Run()
	if err != nil {
		panic(err)
	}

	for {
		bs := make([]byte, 1024)
		n, err := conn.Read(bs)
		if err != nil {
			panic(err)
		} else if n < 7 {
			panic("yeet")
		}
		fmt.Println(string(bs))
	}
}

func pipeOutputToTerminal(label string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(fmt.Sprintf("[%v]"), scanner.Text())
	}
}

func pipeStderrToTerminal(label string, cmd *exec.Cmd) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	defer stderr.Close()
	pipeOutputToTerminal(label, stderr)
}
