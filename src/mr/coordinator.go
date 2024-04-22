package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type JobType int // 0: map, 1: reduce 2: done
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
	Jobchan              chan *Job
	JobchanReducePrapare chan *Job
	Jobtracker           sync.Map
	Worktracker          sync.Map
	Mapnum               int
	ReduceNum            int
	uniJobid             int
	uniWorkerid          int
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

// 向worker发布任务
func (c *Coordinator) GetJob(args *JobRequestArgs, reply *JobRequestReply) error {
	select {
	case job := <-c.Jobchan:
		fmt.Println("get job:", job.Jobid, "job type:", job.JobType)
		reply.Job = job
		reply.Workerid = c.generateWorkerid()
		c.Worktracker.Store(reply.Workerid, 0)
		go c.TrackJob(reply.Workerid, job)
	default:
		reply.Job = nil
	}
	return nil
}

// 追踪worker
func (c *Coordinator) WorkerTracker(args *HreatBeatArgs, reply *HreatBeatReply) error {
	reply.IsReceive = true
	c.Worktracker.Store(args.WorkerId, 1)
	return nil
}

// 检查任务
func (c *Coordinator) JobCheck(args *JobCheckArgs, reply *JobCheckReply) error {
	value, ok := c.Jobtracker.Load(args.Workerid)
	if !ok {
		reply.IsAllow = false
		return nil
	}
	if value.(*Job).Jobid == args.Jobid {
		reply.IsAllow = true
		return nil
	}
	reply.IsAllow = false
	return nil

}

// 追踪任务
func (c *Coordinator) TrackJob(workerid int, job *Job) error {
	c.Jobtracker.Store(workerid, job)
	timer := 0
	for {
		time.Sleep(1 * time.Second)
		if IsDebug {
			fmt.Println("track job:", job.Jobid, "job Type:", job.JobType)
		}
		value, ok := c.Worktracker.Load(workerid)
		if !ok {
			fmt.Println("workerid", workerid, "not found", "tacker job stop!")
			return nil
		}
		if value != 0 {
			c.Worktracker.Store(workerid, 0)
			timer = 0
			continue
		}
		timer++
		if timer > 10 {
			fmt.Println("workerid", workerid, "jobid", job.Jobid, "timeout")
			c.Jobtracker.Delete(workerid)
			c.Worktracker.Delete(workerid)
			c.Jobchan <- job
			return nil
		}
	}
}

// 完成任务
func (c *Coordinator) JobDone(args *JobDoneArgs, reply *JobDoneReply) error {
	c.Worktracker.Delete(args.Workerid)
	c.Jobtracker.Delete(args.Workerid)
	if IsDebug {
		fmt.Println("job done", args.Job.Jobid, args.Workerid)
	}
	switch args.Job.JobType {
	case MapJob:
		c.mapwg.Done()
		c.mu.Lock()
		c.Mapnum--
		Mn := c.Mapnum
		c.mu.Unlock()
		if IsDebug {
			fmt.Println("receive files", args.Job.FileName)
		}
		for i, fn := range args.Job.FileName {
			if i == 0 {
				continue
			}
			TempJob := <-c.JobchanReducePrapare
			TempJob.FileName = append(TempJob.FileName, fn)
			if IsDebug {
				fmt.Println("reduce job", TempJob.Jobid, fn)
			}
			c.JobchanReducePrapare <- TempJob
		}
		if Mn == 0 {
			go c.SendReduceJob()
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

// 任务失败
func (c *Coordinator) JobFailed(args *JobDoneArgs, reply *JobDoneReply) error {
	//to do
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
			c.Jobchan <- job
		}(fp)
	}
}

// 生成jobid
func (c *Coordinator) generateJobid() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.uniJobid += 1
	return c.uniJobid
}

// 生成workerid
func (c *Coordinator) generateWorkerid() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.uniWorkerid += 1
	return c.uniWorkerid

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
			if IsDebug {
				fmt.Println("send reduce job", job.Jobid, job.FileName)
			}
			c.Jobchan <- job
		default:
			return
		}
	}
}

// 发送done任务
func (c *Coordinator) SendDoneJob() {
	c.redwg.Wait()
	c.Jobchan <- &Job{
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
	c.Jobchan = make(chan *Job)
	c.JobchanReducePrapare = make(chan *Job, 10)
	c.Jobtracker = sync.Map{}
	c.Worktracker = sync.Map{}
	c.uniJobid = 0
	c.Mapnum = 0
	c.ReduceNum = 0
	c.uniJobid = 0
	c.mu = sync.Mutex{}
	c.mapwg = sync.WaitGroup{}
	c.redwg = sync.WaitGroup{}
	c.redpwg = sync.WaitGroup{}
	c.isDone = false
	//启动rpc服务
	c.server()
	//制作map任务
	go c.MakeMapJob(files)
	//制作reduce任务
	go c.MakeReduceJob(nReduce)
	return &c
}
