package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type Task struct {
	ID                     uint64      `json:"id,omitempty"`       // my task ID
	Duration               uint64      `json:"duration,omitempty"` // the time to measure the ib_write_bw run for in Seconds
	QP                     uint64      `json:"qp,omitempty"`       // the number of queue pairs to use
	MsgSize                uint64      `json:"msgsize,omitempty"`  // the size of the message to send in bytes
	IgnoreCPUSpeedWarnings bool        `json:"ignorecpuspeedarnings,omitempty"`
	OutputChannel          OutputQueue `json:"-"`
}
type TaskResult struct {
	ServerPort int
	Error      error
}
type JsonErrorResponse struct {
	ErrorText  string
	StatusCode int
}
type JsonResponse struct {
	TaskData    Task
	ServerPort  int
	MessageText string
	StatusCode  int
}

var taskQueue = make(chan Task) // Unbuffered channel for task processing
var maxNics int                 // Number of available nics
var wg sync.WaitGroup
var taskID uint64 = 0 // counter for all jobs
var NicList []string

type OutputQueue chan TaskResult // creat a type to

func parseBodyToTask(r *http.Request) (Task, error) {
	// converts json body to Task struct applying sane defaults if any are missing
	// initilise myTask with the following defaults
	myTask := Task{
		ID:                     0,
		Duration:               5,
		QP:                     2,
		MsgSize:                8383608,
		IgnoreCPUSpeedWarnings: true,
	}
	// the next part reads in json body data and applies any appropriate data to the struct
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&myTask)
	return myTask, err
}

func taskWorker(index int) {
	defer wg.Done()
	for task := range taskQueue {
		fmt.Printf("Processing Task ID: %d from taskworker %d\n", task.ID, index)
		// Simulate task processing time
		//
		err := startIBWriteBW(task, index)
		if err != nil {
			fmt.Printf("in taskworker %v\n", err)
			if err.Error() != "Client unexpectedly exited" {
				// this means something failed during job init like impossible args and we  need to pass it back to the submitTask function below so
				// if can notify the client
				task.OutputChannel <- TaskResult{Error: err}
			}
		}
		fmt.Printf("Task ID %d completed\n", task.ID)
	}
}

func submitTask(w http.ResponseWriter, r *http.Request) {

	// Create a new task and add it to the queue
	// set the task id to
	task, err := parseBodyToTask(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("{\"StatusCode\": 2, \"ErrorText\": \"Failed to Parse the Body you sent as JSON. You sent: '%s'\"}", r.Body), http.StatusBadRequest)
		return
	}
	//var output OutputQueue
	output := make(chan TaskResult)
	task.ID = taskID
	task.OutputChannel = output
	select {
	case taskQueue <- task:
		taskID = taskID + 1
		result := <-output
		if result.Error != nil {
			response := JsonErrorResponse{ErrorText: fmt.Sprintf("An error occured in the server side: %s", result.Error), StatusCode: 1}
			jsonString, err := json.Marshal(response)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(jsonString)
			} else {
				http.Error(w, fmt.Sprintf("{\"StatusCode\": 666, \"ErrorText\": \"trying to format the error response into json this should not happen\"}"), http.StatusInternalServerError)
			}
			return
		} else {
			response := JsonResponse{TaskData: task, MessageText: "started an ib_write_bw process sucessfully", StatusCode: 0, ServerPort: result.ServerPort}
			w.Header().Set("Content-Type", "application/json")
			jsonString, _ := json.Marshal(response)
			w.Write(jsonString)

		}
	default:
		// If channel is full or there are too many tasks running, reject new task
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "{\"StatusCode\": 5, \"ErrorText\": \"Try again later all NICs in use\"}", http.StatusTooManyRequests)
	}
}

func main() {
	NicList = []string{"bnxt_re0", "bnxt_re1", "bnxt_re2", "bnxt_re3"}
	maxNics = len(NicList)
	// Start workers (equal to the number of available nics)
	for i := 0; i < maxNics; i++ {
		wg.Add(1)
		go taskWorker(i)
	}

	// HTTP handler for submitting tasks
	http.HandleFunc("/work", submitTask)

	// Start HTTP server
	port := 8000

	fmt.Printf("Starting server on port %v...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	// Wait for all workers to finish before exiting
	wg.Wait()
}
