package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Any interface{}

// Queue is a basic FIFO Queue
type Queue struct {
	nodes []Any
	size  int
	head  int
	tail  int
	count int
}

// Push adds a item to the Queue.
func (q *Queue) Push(n Any) {
	// The queue size is reached it doesn't resize. It just pops the item.
	if q.count == q.size {
		q.Pop()
	}
	q.nodes[q.backIndex()] = n
	q.count++
}

// Pop removes and returns the item.
func (q *Queue) Pop() Any {
	if q.count == 0 {
		return nil
	}
	item := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return item
}

// queue length
func (q *Queue) Len() int {
	return q.count
}

func (q *Queue) backIndex() int {
	return (q.head + q.count) % q.size
}

// http response status
type Status struct {
	Url          string
	TimeOfTheDay time.Time
	StatuCode    int
	Error        error
}

// creates a new instance of the queue
func NewQueue(size int) *Queue {
	return &Queue{
		nodes: make([]Any, size),
		size:  size,
	}
}

const (
	maxFailures = 3               // Max failures after which someone has to be notified.
	duration    = time.Second * 5 // how often the site should be polled.
)

func main() {

	tweet := fmt.Println

	// ticker channel which posts the 5 seconds
	timer := time.NewTicker(duration)

	// Failure statusChannel channel
	statusChannel := make(chan Status, 10)

	// tweetfailureChannel of the failuers
	tweetfailureChannel := make(chan string)

	url := "http://www.google1.com"

	// Checks the site Status.Ticker goroutine
	go func() {
		for t := range timer.C {
			_ = t
			resp, err := http.Get(url)
			if err != nil {
				statusChannel <- Status{Url: url, TimeOfTheDay: time.Now(), Error: err}
			} else if resp.StatusCode != 200 {
				statusChannel <- Status{url, time.Now(), resp.StatusCode, nil}
			}
		}
	}()

	// goroutine to check number of failures within specified duration.
	go func() {
		queue := NewQueue(maxFailures)
		for response := range statusChannel {
			if queue.Len() == maxFailures {
				if queue.Pop().(Status).TimeOfTheDay.Sub(response.TimeOfTheDay).Seconds() < 60 {
					tweetfailureChannel <- "Couldnt reach " + response.Url + " " + strconv.Itoa(maxFailures) + " times within the last 60 seconds"
					timer.Stop()
				}
			} else {
				queue.Push(response)
			}
		}
	}()

	//
	go func() {
		for message := range tweetfailureChannel {
			tweet(message)
		}
	}()
	time.Sleep(time.Minute * 10)
}
