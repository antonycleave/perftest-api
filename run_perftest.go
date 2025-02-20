package main

import (
/*	"log"
	"os"
	"os/exec" */
	"fmt"
	"strconv"
	"net"
	"time"
)

func buildIBWriteBWArgs(myTask Task) ([]string) {
	//this will convert the supplied args inti
    var arglist []string
	arglist = append(arglist,"--duration", strconv.FormatUint(myTask.Duration, 10))
	arglist = append(arglist, "-q", strconv.FormatUint(myTask.QP, 10))
	arglist = append(arglist, "-s", strconv.FormatUint(myTask.MsgSize, 10))
	if myTask.IgnoreCPUSpeedWarnings {
		arglist = append(arglist, "-F")
	}
	return arglist
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


func startIBWriteBW(myTask Task, nicIndex int) {
    tcpPort, _ := Listener()
	arglist := buildIBWriteBWArgs(myTask)
    arglist = append([]string{"/opt/perftest-with-rocm/bin/ib_write_bw", "-p", fmt.Sprintf("%d", tcpPort), "-d", fmt.Sprintf("%s", NicList[nicIndex])}, arglist...)
	fmt.Println(arglist)
	// this will start an ib_write_bw_process
	/*
    cmd := exec.Command( arglist )
    err := cmd.Start()
    time.Sleep(time.Duration(myTask.Duration) * time.Second)
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
	time.Sleep(time.Duration(myTask.Duration) * time.Second)
    return

}