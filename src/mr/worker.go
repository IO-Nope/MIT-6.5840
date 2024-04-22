package mr

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"time"
)

var IsDebug bool = false
var Issh bool = true

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

type WriteType struct {
	Kvs  []KeyValue
	Path string
}
type ReadType struct {
	Path string
	Kvs  []KeyValue
}

func Check(e error) {
	if !IsDebug {
		return
	}
	if e != nil {
		fmt.Println("error! :", e)
	}
}

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	// 初始化
	for {
		func() {
			job, workerid := JobRequest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go AliveBeats(workerid, ctx)
			if job == nil {
				if IsDebug {
					fmt.Println("no job")
				}
				time.Sleep(1 * time.Second)
				return
			}
			if IsDebug {
				fmt.Println("jobtype:", job.JobType, "jobfiles:", job.FileName, "jobid:", job.Jobid)
			}
			switch job.JobType {
			case MapJob:
				intermediate := []KeyValue{}
				Rq := ReadRequest(workerid, job.Jobid)
				if !Rq {
					return
				}
				path := ""
				if Issh {
					path = job.FileName[0]
				} else {
					path = "../main/" + job.FileName[0]
				}
				file, err := os.Open(path)
				if IsDebug {
					fmt.Println(job.FileName[0])
				}
				if err != nil {
					fmt.Println(err)
					return
				}
				content, err := io.ReadAll(file)
				if err != nil {
					fmt.Println(err)
					return
				}
				file.Close()
				kva := mapf(job.FileName[0], string(content))
				intermediate = append(intermediate, kva...)
				// 写入json
				Wq := WriteRequest(workerid, job.Jobid)
				if !Wq {
					return
				}
				for _, kv := range intermediate {
					path := ""
					if Issh {
						path = "../" + "mr-" + strconv.Itoa(job.Jobid) + "-" + strconv.Itoa(ihash(kv.Key)%10) + ".json"
					} else {
						path = "../main/" + "mr-" + strconv.Itoa(job.Jobid) + "-" + strconv.Itoa(ihash(kv.Key)%10) + ".json"
					}
					writedone := WriteType{Kvs: []KeyValue{kv}, Path: path}
					WriteKvsToJson(writedone)
					fs := "mr-" + strconv.Itoa(job.Jobid) + "-" + strconv.Itoa(ihash(kv.Key)%10) + ".json"
					for i, f := range job.FileName {
						if f == fs {
							break
						}
						if i == len(job.FileName)-1 {
							job.FileName = append(job.FileName, fs)
						}
					}

				}
				SendJobDone(job, workerid)
			case ReduceJob:
				// reduce
				intermediate := []KeyValue{}
				Rq := ReadRequest(workerid, job.Jobid)
				if !Rq {
					return
				}
				for _, filename := range job.FileName {
					path := ""
					if Issh {
						path = "../" + filename
					} else {
						path = "../main/" + filename
					}
					kvs := ReadKvsFromJson(path)
					if kvs == nil {
						if IsDebug {
							fmt.Println("kvs is nil")
						}
						return
					}
					intermediate = append(intermediate, kvs...)
				}
				sort.Sort(ByKey(intermediate))
				oname := ""
				if Issh {
					oname = "mr-out-" + strconv.Itoa(job.Reducenum) + ".txt"
				} else {
					oname = "../main/" + "mr-out-" + strconv.Itoa(job.Reducenum) + ".txt"
				}
				Wq := WriteRequest(workerid, job.Jobid)
				if !Wq {
					return
				}
				ofile, _ := os.Create(oname)
				i := 0
				for i < len(intermediate) {
					j := i + 1
					for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
						j++
					}
					values := []string{}
					for k := i; k < j; k++ {
						values = append(values, intermediate[k].Value)
					}
					output := reducef(intermediate[i].Key, values)

					// this is the correct format for each line of Reduce output.
					fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

					i = j
				}
				ofile.Close()
				SendJobDone(job, workerid)
			case Done:
				return
			}
		}()
	}
}

// uncomment to send the Example RPC to the coordinator.
// CallExample()
// 一个用来debug的func
func Debugfunc(vals ...interface{}) {}

// 将kv对写入.json
func WriteKvsToJson(kvswtype WriteType) {
	file, err := os.OpenFile(kvswtype.Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	for _, kv := range kvswtype.Kvs {
		err := enc.Encode(&kv)
		if err != nil {
			fmt.Println(err)
			return
		}
		if err != nil {
			break
		}
	}
}

// 读取请求
func ReadRequest(workerid int, jobid int) bool {
	args := JobCheckArgs{Jobid: jobid, Workerid: workerid}
	reply := JobCheckReply{}
	ok := call("Coordinator.JobCheck", &args, &reply)
	if !ok {
		return false
	}
	return reply.IsAllow
}

// 写入请求
func WriteRequest(workerid int, jobid int) bool {
	args := JobCheckArgs{Jobid: jobid, Workerid: workerid}
	reply := JobCheckReply{}
	ok := call("Coordinator.JobCheck", &args, &reply)
	if !ok {
		return false
	}
	return reply.IsAllow
}

// 从.json读取kv对
func ReadKvsFromJson(filepath string) []KeyValue {
	file, err := os.Open(filepath)
	if err != nil {
		if IsDebug {
			fmt.Fprintln(os.Stderr, "open file error", err, filepath)
		}
		return nil
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	kva := []KeyValue{}
	for {
		var kv KeyValue
		if err := dec.Decode(&kv); err != nil {
			break
		}
		kva = append(kva, kv)
	}

	return kva
}
func SendJobDone(Job *Job, workerid int) {
	args := JobDoneArgs{Job: Job, Workerid: workerid}
	reply := JobDoneReply{}
	ok := call("Coordinator.JobDone", &args, &reply)
	if ok {
		if IsDebug {
			fmt.Println("JobDone! jobtype:", Job.JobType, "jobid:", Job.Jobid, "workerid:", workerid)
		}
	}
}
func JobRequest() (*Job, int) {
	args := JobRequestArgs{AskJob: true}
	reply := JobRequestReply{}
	ok := call("Coordinator.GetJob", &args, &reply)
	if ok {
		//fmt.Println("reply.Y", reply.Job.jobid)
		return reply.Job, reply.Workerid
	}
	return nil, -1
}
func AliveBeats(workerId int, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
			args := HreatBeatArgs{WorkerId: workerId}
			reply := HreatBeatReply{}
			ok := call("Coordinator.WorkerTracker", &args, &reply)
			if ok {
				if IsDebug {
					fmt.Println("reply.IsReceive", reply.IsReceive)
				}
			}
		}
	}
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
