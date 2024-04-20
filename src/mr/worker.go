package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

var initid int = 0
var IsDebug bool = true

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
	for {
		Job := JobRequest(initid)
		initid += 1
		if Job == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if Job == nil {
			break

		}
		if IsDebug {
			fmt.Println(Job.JobType, Job.FileName, Job.Reducenum, Job.Jobid)
		}
		switch Job.JobType {
		case MapJob:
			go func() {
				intermediate := []KeyValue{}
				file, err := os.Open("../main/" + Job.FileName[0])
				if IsDebug {
					fmt.Println(Job.FileName[0])
				}
				Check(err)
				content, err := io.ReadAll(file)
				Check(err)
				file.Close()
				kva := mapf(Job.FileName[0], string(content))
				intermediate = append(intermediate, kva...)
				path := "../main/" + "mr-" + strconv.Itoa(Job.Jobid) + "-" + strconv.Itoa(ihash(Job.FileName[0])) + ".json"
				WriteKvsToJson(intermediate, path)
				SendMapJobDone(Job)
			}()
		case ReduceJob:
			// reduce
		default:
			break
		}
	}
	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}
func Debugfunc(vals ...interface{}) {

}
func WriteKvsToJson(kvs []KeyValue, filepath string) {
	file, err := os.Create(filepath)
	Check(err)
	defer file.Close()
	enc := json.NewEncoder(file)
	for _, kv := range kvs {
		err := enc.Encode(&kv)
		Check(err)
		if err != nil {
			break
		}
	}
}
func SendMapJobDone(Job *Job) {

}
func JobRequest(workerId int) *Job {
	args := JobRequestArgs{WorkerId: workerId}
	reply := JobRequestReply{}
	ok := call("Coordinator.GetJob", &args, &reply)
	if ok {
		//fmt.Println("reply.Y", reply.Job.jobid)
		return reply.Job
	}
	return nil
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
