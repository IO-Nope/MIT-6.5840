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
	MapJob    JobType = iota
	ReduceJob JobType = iota
	Done
)

type Job struct {
	JobType   JobType
	FileName  []string
	Reducenum int
	Jobid     int
}

type Coordinator struct {
	// Your definitions here.
	JobchanMap           chan *Job
	JobchanReduce        chan *Job
	JobchanReducePrapare chan *Job
	Mapnum               int
	ReduceNum            int
	uniJobid             int
	mu                   sync.Mutex
	isDone               bool
	mapwg                sync.WaitGroup
	redwg                sync.WaitGroup
	redpwg               sync.WaitGroup
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
	}
	return nil
}

// 完成任务
func (c *Coordinator) JobDone(args *JobDoneArgs, reply *JobDoneReply) error {
	switch args.Job.JobType {
	case MapJob:
		c.mapwg.Done()
		c.mu.Lock()
		c.Mapnum--
		Mn := c.Mapnum
		c.mu.Unlock()
		fmt.Println("receive files", args.Job.FileName)
		for i, fn := range args.Job.FileName {
			if i == 0 {
				continue
			}
			TempJob := <-c.JobchanReducePrapare
			c.mu.Lock()
			TempJob.FileName = append(TempJob.FileName, fn)
			if IsDebug {
				fmt.Println("reduce job", TempJob.Jobid, fn)
			}
			c.mu.Unlock()
			c.JobchanReducePrapare <- TempJob
		}
		if IsDebug {
			fmt.Println("mapnum", Mn)
		}
		if Mn == 0 {
			c.SendReduceJob()
		}
	case ReduceJob:
		c.redwg.Done()
		for _, fn := range args.Job.FileName {
			path := "../main/" + fn
			err := os.Remove(path)
			if IsDebug {
				fmt.Println("remove file", path, err)
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

// 制作map任务
func (c *Coordinator) MakeMapJob(files []string) {
	for _, fp := range files {
		c.mapwg.Add(1)
		c.mu.Lock()
		c.Mapnum++
		c.mu.Unlock()
		go func(fp string) {
			id := c.generateJobid()
			fmt.Println("making map job", id)
			job := &Job{
				JobType:   MapJob,
				FileName:  []string{fp},
				Reducenum: 0,
				Jobid:     id,
			}
			fmt.Println("jobid", id)
			c.JobchanMap <- job
		}(fp)
	}
}
func (c *Coordinator) generateJobid() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.uniJobid += 1
	fmt.Println(c.uniJobid)
	return c.uniJobid
}

// 制作reduce任务
func (c *Coordinator) MakeReduceJob(nReduce int) {
	//to do
	for i := 0; i < nReduce; i++ {
		c.redwg.Add(1)
		go func(i int) {
			id := c.generateJobid()
			fmt.Println("making reduce job", id)
			c.mu.Lock()
			RN := c.ReduceNum
			c.ReduceNum++
			c.mu.Unlock()
			job := &Job{
				JobType:   ReduceJob,
				FileName:  []string{},
				Reducenum: RN,
				Jobid:     id,
			}
			c.JobchanReducePrapare <- job
		}(i)

	}
	go c.SendDoneJob()
}

// 发送reduce任务
func (c *Coordinator) SendReduceJob() {
	for {
		select {
		case job := <-c.JobchanReducePrapare:
			fmt.Println("send reduce job", job.Jobid, job.FileName)
			c.JobchanReduce <- job
		default:
			return
		}
	}
}

// 发送done任务
func (c *Coordinator) SendDoneJob() {
	c.redwg.Wait()
	c.JobchanReduce <- &Job{
		JobType:   Done,
		FileName:  []string{},
		Reducenum: 0,
		Jobid:     c.generateJobid(),
	}
	c.isDone = true
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

// 初始化coordinator
func (c *Coordinator) Initialization() {

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
	c.JobchanReducePrapare = make(chan *Job, 10)
	c.uniJobid = 0
	c.Mapnum = 0
	c.ReduceNum = 0
	c.mu = sync.Mutex{}
	c.mapwg = sync.WaitGroup{}
	c.redwg = sync.WaitGroup{}
	c.redpwg = sync.WaitGroup{}
	c.isDone = false
	//启动rpc服务
	c.server()
	//制作map任务
	go c.MakeMapJob(files)
	go c.MakeReduceJob(nReduce)
	return &c
}
