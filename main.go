package main

import (
	"fmt"
	"net/http"
	"os"
//	"strconv"
	"sync"
	"encoding/json"
)

type Task struct {
	ID uint64 `json:"id,omitempty"` // my task ID
	Duration uint `json:"Duration,omitempty"` // the time to measure the ib_write_bw run for in Seconds
	QP uint `json:"QP,omitempty"`// the number of queue pairs to use
	MsgSize uint `json:"MsgSize,omitempty"`// the size of the message to send in bytes
	IgnoreCPUSpeedWarnings bool `json:"IgnoreCPUSpeedWarnings,omitempty"`
}

func  parseBodyToTask(body []byte, id uint64) (Task, error) {
    // converts json body to Task struct applying sane defaults if any are missing
	// initilise myTask with the following defaults
	myTask := Task {
		ID: id,
        Duration: 20,
		QP: 2,
		MsgSize: 8383608,
		IgnoreCPUSpeedWarnings: true }
	// the next part reads in json body data and applies any appropriate data to the struct
	err := json.Unmarshal(body, &myTask)
	// if some sneaky bugger tries to set their own id spoil their fun
	myTask.ID=id
	return myTask, err
}

var taskQueue = make(chan Task) // Unbuffered channel for task processing
var maxNics = 4                 // Number of available nics
var wg sync.WaitGroup
var taskID uint64 = 4 // counter for all jobs
func taskWorker() {
	defer wg.Done()
	for task := range taskQueue {

		fmt.Printf("Processing Task ID: %d\n", task.ID)
		// Simulate task processing time
		//
		startIBWriteBW(task)
		fmt.Printf("Task ID %d completed\n", task.ID)
	}
}

func submitTask(w http.ResponseWriter, r *http.Request) {
	// Create a new task and add it to the queue
	var body = []byte(`{"QP": 3}`)
	task, err := parseBodyToTask( body, taskID )
    if err != nil {
		fmt.Println("err\n",err)
	}
	select {
	case taskQueue <- task:
		fmt.Fprintf(w, "Task ID %d added to queue\n", task.ID)
		taskID = taskID + 1
	default:
		// If channel is full or there are too many tasks running, reject new task
		http.Error(w, "Server is busy, try again later", http.StatusTooManyRequests)
	}
}

func main() {
	// Start workers (equal to the number of available nics)
	for i := 0; i < maxNics; i++ {
		//wg.Add(1)
		go taskWorker()
	}

	// HTTP handler for submitting tasks
	http.HandleFunc("/work", submitTask)

	// Start HTTP server
	port := 8000

	fmt.Printf("Starting server on port %v...", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	// Wait for all workers to finish before exiting
	//wg.Wait()
}
