package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const ibWriteClientWait uint64 = 20

/*
ibWriteErrorWait is the time to wait to ensure that the server process has started
it also is used as the delay between default checks
*/
const ibWriteErrorWait = 100 * time.Millisecond

func buildIBWriteBWArgs(myTask Task) []string {
	//this will convert the supplied args inti
	var arglist []string
	arglist = append(arglist, "--duration", strconv.FormatUint(myTask.Duration, 10))
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

func startIBWriteBW(myTask Task, nicIndex int) error {
	tcpPort, _ := Listener()
	arglist := buildIBWriteBWArgs(myTask)

	arglist = append([]string{"-p", fmt.Sprintf("%d", tcpPort), "-d", fmt.Sprintf("%s", NicList[nicIndex])}, arglist...)
	//arglist = append([]string{"-p", fmt.Sprintf("%d", tcpPort), "-d", "bnxt_re9"}, arglist...)
	cmd := "/opt/perftest-with-rocm/bin/ib_write_bw"
	//cmd = "sleep"
	//arglist= []string{"36"}
	fmt.Printf("running %s %s\n", cmd, strings.Join(arglist, " "))

	// this will start an ib_write_bw server process
	ib_write_bw_cmd := exec.Command(cmd, arglist...)
	ib_write_bw_cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// thanks stack overflow! https://stackoverflow.com/questions/32921792/how-do-you-kill-a-process-and-its-children-on-a-timeout-in-go-code
	// buffered chan is important so the goroutine does't
	// get blocked and stick around if the function returns
	// after the timeout
	type IBWBResult struct {
		CombinedOutput []byte
		Error          error
	}
	done := make(chan IBWBResult, 1)
	go func() {
		output, err := ib_write_bw_cmd.CombinedOutput()
		res := IBWBResult{CombinedOutput: output, Error: err}
		done <- res
	}()
	fmt.Printf("port=%d\n", tcpPort)
	sent_ok := false
	serverStarttime := time.Now()
	for {
		select {
		case ibwbresult := <-done:
			if ibwbresult.Error != nil {
				return fmt.Errorf("%s\n%s\n", ibwbresult.Error, ibwbresult.CombinedOutput)
			}
			return nil
		default:
			if !sent_ok && time.Now().Sub(serverStarttime) > ibWriteErrorWait {
				if ib_write_bw_cmd.ProcessState.ExitCode() == -1 {
					sent_ok = true
					myTask.OutputChannel <- TaskResult{ServerPort: tcpPort}
				}
			} else if time.Now().Sub(serverStarttime) > (time.Duration(myTask.Duration+ibWriteClientWait) * time.Second) {
				// this next bit kills the process and frees up the worker for another job
				pid := ib_write_bw_cmd.Process.Pid
				pgid, err := syscall.Getpgid(pid)
				if err == nil {
					fmt.Printf("Client didn't connect or finish in time. Killing process. The pid is %d, and the pgid we are killing is %d\n", pid, pgid)
					err := syscall.Kill(-pgid, syscall.SIGKILL)
					if err != nil {
						fmt.Printf("kill Command finished with error: %v. the pid is/was %d\n", err, pid)
						return err
					}
				}
				return nil
			}
			time.Sleep(ibWriteErrorWait)
		}
	}
	return nil
}
