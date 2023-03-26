package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

func main() {
	cmd := exec.Command("/bin/bash")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Println("Error starting pty:", err)
		return
	}

	// Set the PTY's window size
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		ptmx.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ Row, Col, Xpixel, Ypixel uint16 }{40, 80, 0, 0})),
	)
	if errno != 0 {
		fmt.Println("Error setting window size:", errno)
		return
	}

	// Catch the interrupt signal and send the ctrl+c key sequence to the PTY
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for {
			<-sigChan
			_, _ = ptmx.WriteString("\x03")
		}
	}()

	// Copy input/output to the PTY
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	_, _ = io.Copy(os.Stdout, ptmx)

	_ = cmd.Wait()
}
