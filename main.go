package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	command := exec.Command("../timedecho/timedecho")
	dataFromProcess, sendDataToProcess := execWithChannel(command)

	go sendingData(sendDataToProcess)

	go func() {
		for {
			select {
			case data := <-dataFromProcess:
				fmt.Println("Did not time out, output: ", data)
			case <-time.After(time.Second * 1):
				fmt.Println("Timeout after waiting new line for 1 second, killing process")
				command.Process.Signal(syscall.SIGKILL)
			}
		}
	}()

	command.Wait()
}

func sendingData(sendDataToProcess chan interface{}) {
	for i := 1; i <= 3; i++ {
		sendDataToProcess <- i * 500000
	}
	sendDataToProcess <- nil
}

func sendData(writterToCmdStdin *io.PipeWriter) {

	writterToCmdStdin.Close()
}

func pipeReaderToChannel(pipeReader *io.PipeReader) chan string {
	channel := make(chan string)
	scanner := bufio.NewScanner(pipeReader)
	go func() {
		for scanner.Scan() {
			channel <- scanner.Text()
		}
	}()

	return channel
}

func pipeWriterToChannel(pipWriter *io.PipeWriter) chan interface{} {
	channel := make(chan interface{})
	go func() {
		for {
			data := <-channel
			if data == nil {
				pipWriter.Close()
			} else {
				fmt.Fprintln(pipWriter, data)
			}
		}
	}()

	return channel
}

func execWithChannel(command *exec.Cmd) (chan string, chan interface{}) {
	stdin, writerToCmdStdin := io.Pipe()
	readerToCmdStdout, stdout := io.Pipe()
	command.Stdin = stdin
	command.Stdout = stdout
	command.Start()

	return pipeReaderToChannel(readerToCmdStdout), pipeWriterToChannel(writerToCmdStdin)
}
