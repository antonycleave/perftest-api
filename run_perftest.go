package main

import (
/*	"log"
	"os"
	"os/exec" */
	"net"
	"time"
)

func buildIBWriteBWArgs() {
	//this will convert the supplied args inti

}

func Listener() (port int, err error) {
	if a, err := net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		if l, err := net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}


func startIBWriteBW() {
	buildIBWriteBWArgs()
	// this will start an ib_write_bw_process
    /*cmd := exec.Command( "" )
    err := cmd.Start()
    if err != nil {
        return err
    }
    pid := cmd.Process.Pid
    // use goroutine waiting, manage process
    // this is important, otherwise the process becomes in S mode
    go func() {
        err = cmd.Wait()
        fmt.Printf("Command finished with error: %v", err)
    }()
    */
	time.Sleep(2 * time.Second)
    return

}