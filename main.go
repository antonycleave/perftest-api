package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Task struct {
	ID int
}

var taskQueue = make(chan Task) // Unbuffered channel for task processing
var maxNics = 4 // Number of available nics
var wg sync.WaitGroup

func taskWorker() {
	defer wg.Done()
	for task := range taskQueue {
		fmt.Printf("Processing Task ID: %d\n", task.ID)
		// Simulate task processing time
		time.Sleep(2 * time.Second)
		fmt.Printf("Task ID %d completed\n", task.ID)
	}
}

func submitTask(w http.ResponseWriter, r *http.Request, id int) {
	fmt.Fprintf(w, "There are %d jobs in the queue\n", len(taskQueue))
	if len(taskQueue) >= maxNics {
		http.Error(w, "Too many tasks in the queue, try again later", http.StatusTooManyRequests)
		return
	}

	// Create a new task and add it to the queue
	task := Task{ID: id}

	select {
	case taskQueue <- task:
		fmt.Fprintf(w, "Task ID %d added to queue\n", taskID)
	default:
		// If channel is full or there are too many tasks running, reject new task
		http.Error(w, "Server is busy, try again later", http.StatusServiceUnavailable)
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
	fmt.Println("Starting server on port 8000...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	// Wait for all workers to finish before exiting
	//wg.Wait()
}
