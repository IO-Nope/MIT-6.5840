package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	pf "path/filepath"
	"sort"
	"strconv"
)

var IsDebug bool = true

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

	job := JobRequest(0)
	if job == nil {
		if IsDebug {
			fmt.Println("no job")
		}
		return
	}
	if IsDebug {
		fmt.Println(job.JobType, job.FileName, job.Jobid)
	}
	switch job.JobType {
	case MapJob:
		intermediate := []KeyValue{}
		file, err := os.Open("../main/" + job.FileName[0])
		if IsDebug {
			fmt.Println(job.FileName[0])
		}
		Check(err)
		content, err := io.ReadAll(file)
		Check(err)
		file.Close()
		kva := mapf(job.FileName[0], string(content))
		intermediate = append(intermediate, kva...)
		for _, kv := range intermediate {
			path := "../main/" + "mr-" + strconv.Itoa(job.Jobid) + "-" + strconv.Itoa(ihash(kv.Key)%10) + ".json"
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
		SendJobDone(job, 0)
	case ReduceJob:
		// reduce
		intermediate := []KeyValue{}
		for _, filename := range job.FileName {
			path := "../main/" + filename
			kvs := ReadKvsFromJson(path)
			intermediate = append(intermediate, kvs...)
		}
		sort.Sort(ByKey(intermediate))
		oname := "../main/" + "mr-out-" + strconv.Itoa(job.Reducenum) + ".txt"
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
		SendJobDone(job, 0)
	case Done:
		return
	}
}

// uncomment to send the Example RPC to the coordinator.
// CallExample()

func Debugfunc(vals ...interface{}) {

}
func WriteKvsToJson(kvswtype WriteType) {
	file, err := os.OpenFile(kvswtype.Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	Check(err)
	defer file.Close()
	enc := json.NewEncoder(file)
	for _, kv := range kvswtype.Kvs {
		err := enc.Encode(&kv)
		Check(err)
		if err != nil {
			break
		}
	}
}
func AimFiles(filepath string, aimnum int) []string {
	file, err := os.ReadDir(filepath)
	Check(err)
	Filesname := []string{}
	for _, f := range file {
		if f.Type().IsDir() {
			continue
		}
		if pf.Ext(filepath+"/"+f.Name()) != ".json" {
			continue
		}
		if f.Name()[len(f.Name())-6] == strconv.Itoa(aimnum)[0] {
			Filesname = append(Filesname, filepath+"/"+f.Name())
		}
	}
	return Filesname
}
func ReadTrigger() {

}
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
			fmt.Println("reply.IsReceive", Job.FileName, Job.Jobid, workerid)
		}
	}
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
