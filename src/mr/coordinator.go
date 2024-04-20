package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

type JobType int // 0: map, 1: reduce
const (
	MapJob JobType = iota
	ReduceJob
)

type Job struct {
	JobType   JobType
	FileName  []string
	Reducenum int
	Jobid     int
}

type Coordinator struct {
	// Your definitions here.
	JobchanMap    chan *Job
	JobchanReduce chan *Job
	Mapnum        int
	ReduceNum     int
	uniJobid      int
	mu            sync.Mutex
	isDone        bool
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 2
	return nil
}
func (c *Coordinator) GetJob(args *JobRequestArgs, reply *JobRequestReply) error {
	select {
	case job := <-c.JobchanMap:
		fmt.Println("get map job", job.Jobid)
		reply.Job = job
	case job := <-c.JobchanReduce:
		fmt.Println("get reduce job", job.Jobid)
		reply.Job = job
	default:
		reply.Job = nil
		c.mu.Lock()
		c.isDone = true
		c.mu.Unlock()
	}
	return nil
}

// 制作map任务
func (c *Coordinator) MakeMapJob(files []string) {
	for _, fp := range files {
		go func(fp string) {
			id := c.generateJobid()
			fmt.Println("making map job", id)
			job := &Job{
				JobType:   MapJob,
				FileName:  []string{fp},
				Reducenum: c.ReduceNum,
				Jobid:     id,
			}
			fmt.Println("jobid", id)
			c.JobchanMap <- job
			c.mu.Lock()
			c.Mapnum++
			c.mu.Unlock()
		}(fp)
	}
	fmt.Println("done making map jobs")
}
func (c *Coordinator) generateJobid() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.uniJobid += 1
	fmt.Println(c.uniJobid)
	return c.uniJobid
}

// 制作reduce任务
func (c *Coordinator) MakeReduceJob() {
	//to do
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
	c.mu.Lock()
	defer c.mu.Unlock()
	// Your code here.
	return c.isDone
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	//初始化coordinator
	c.JobchanMap = make(chan *Job)
	c.JobchanReduce = make(chan *Job)
	c.uniJobid = 0
	c.Mapnum = 0
	c.ReduceNum = nReduce
	c.mu = sync.Mutex{}
	c.isDone = false
	//启动rpc服务
	c.server()
	//制作map任务
	go c.MakeMapJob(files)
	go c.MakeReduceJob()
	return &c
}
