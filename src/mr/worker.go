package mr

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

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
	//some preparation
	fmt.Println("Worker started")
	// intermediate := []KeyValue{}
	go func() {
		for {

			filename := MapCall()
			go func() {
				filename = "../main/" + filename
				intermediate := []KeyValue{}
				file, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				content, err := io.ReadAll(file)
				if err != nil {
					log.Fatalf("cannot read %v", filename)
				}
				file.Close()
				kva := mapf(filename, string(content))
				intermediate = append(intermediate, kva...)
				Emit(intermediate)
			}()
		}
	}()

	//请求线程

	//与coordination通信
	//Map任务 kv

	//reduce任务 返回

	//参考mrsquential.go
	// uncomment to send the Example RPC to the coordinator.

	fmt.Println("Worker finished", imm)
}
func MapCall() string {
	args := MapArgs{}
	args.Routinenum = 1
	reply := MapReply{}
	//fmt.Println(1)
	ok := call("Coordinator.MapHeadlers", &args, &reply)
	//fmt.Println(2)
	if ok {
		//fmt.Printf("reply.Y %v\n", reply.Filename)
		return reply.Filename
	} else {
		fmt.Printf("call failed!\n")
		return ""
	}
}
func EmitM(intermediate []KeyValue) {
	args := EmitMArgs{}
	args.Intermediate = intermediate
	reply := EmitMReply{}
	ok := call("Coordinator.ReduceEmitHeadlers", &args, &reply)
	if ok {
		fmt.Printf("Isdone %v\n", reply.Isdone)
	} else {
		fmt.Printf("call failed!\n")
	}
}
func EmitR() {

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
