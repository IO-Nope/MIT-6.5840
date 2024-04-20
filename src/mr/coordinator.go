package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"

	"6.5840/mytools"
)

type Coordinator struct {
	// Your definitions here.
	sumch      chan string
	files      []string
	ReducePool mytools.WorkerPool
	wg         sync.WaitGroup
	isdone     bool
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) ReduceHeadlers(args *ReduceArgs, reply *ReduceReply) error {

	return nil
}
func (c *Coordinator) MapHeadlers(args *MapArgs, reply *MapReply) error {
	fmt.Println("MapHeadlers")
	reply.Filename = <-c.sumch
	c.wg.Done()
	fmt.Println(reply.Filename)
	return nil
}
func (c *Coordinator) MaperrHeadlers(args *MaperrArgs, reply *MaperrReply) error {
	c.wg.Add(1)
	c.sumch <- args.Filename
	return nil
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	if c.isdone {
		return true
	}
	return false
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	// initialize
	c.ReducePool = *mytools.NewWorkerPool(nReduce)
	c.sumch = make(chan string)
	c.files = files
	go func() {
		for _, file := range c.files {
			fmt.Println(file)
			c.sumch <- file
			c.wg.Add(1)
		}
	}()
	c.server()
	return &c
}
