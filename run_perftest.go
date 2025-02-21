package main

import (
	//"log"
	//"os"
	"syscall"
	"os/exec"
	"fmt"
	"strconv"
	"strings"
	"net"
	"time"
)
const ibWriteClientWait uint64 = 5
const ibWriteErrorWait = 100 * time.Millisecond

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


func startIBWriteBW(myTask Task, nicIndex int) (error) {
    cmd := "/opt/perftest-with-rocm/bin/ib_write_bw"
	tcpPort, _ := Listener()
	arglist := buildIBWriteBWArgs(myTask)

    //arglist = append([]string{"-p", fmt.Sprintf("%d", tcpPort), "-d", fmt.Sprintf("%s", NicList[nicIndex])}, arglist...)
	arglist = append([]string{"-p", fmt.Sprintf("%d", tcpPort), "-d", "bnxt_re9"}, arglist...)
	cmd = "/opt/perftest-with-rocm/bin/ib_write_bw"
	cmd="sleep"
	arglist = []string{"3600"}
	fmt.Printf("running %s \"%s\"\n", cmd, strings.Join(arglist, "\" "))

	// this will start an ib_write_bw server process
    ib_write_bw_cmd := exec.Command(cmd,  arglist... )
	ib_write_bw_cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
    err := ib_write_bw_cmd.Start()
	if err != nil {
        return err
		fmt.Printf("Command finished with error: %v. the pid is/was %d\n", err, ib_write_bw_cmd.Process.Pid)
    }
	// thanks stack overflow! https://stackoverflow.com/questions/32921792/how-do-you-kill-a-process-and-its-children-on-a-timeout-in-go-code
	// buffered chan is important so the goroutine does't
	// get blocked and stick around if the function returns
	// after the timeout
	done := make(chan error, 1)
	    go func() {
		done <- ib_write_bw_cmd.Wait()
	}()
    select {
	case err := <-done:
		// this will be nil if no error
		return err
	case <-time.After(time.Duration(myTask.Duration + ibWriteClientWait) * time.Second):
		pid := ib_write_bw_cmd.Process.Pid
		pgid, err := syscall.Getpgid(pid)
		if err == nil {
			fmt.Printf("Client didn't connect in time. Killing process. The pid is %d, and the pgid we are killing is %d\n", pid, pgid)
			err := syscall.Kill(-pgid, syscall.SIGKILL)
			if err != nil {
				fmt.Printf("kill Command finished with error: %v. the pid is/was %d\n", err, pid)
				return  err
			}
		}
	}
	return nil
}