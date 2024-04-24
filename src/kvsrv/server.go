package kvsrv

import (
	"log"
	"sync"
	"time"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type KVServer struct {
	mu      sync.Mutex
	log     sync.Map // 用于存储日志 int bool
	kvsaver sync.Map // 用于存储kv string string
	// taskchan     chan chan int64
	//getlog       sync.Map
	serverticker time.Time
	// Your definitions here.
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	DPrintf("Get request :%v \n at time: %s", args, kv.serverticker.Format(time.RFC3339Nano))
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if !args.Logmode {
		kv.log.Delete(args.Id)
		// kv.getlog.Delete(args.Id)
		return
	}
	// task := make(chan int64)
	// kv.taskchan <- task
	// task <- args.Id
	// if value, ok := kv.getlog.Load(args.Id); !ok {
	if value, ok := kv.kvsaver.Load(args.Key); ok {
		reply.Value = value.(string)
	} else {
		reply.Value = ""
	}
	// kv.getlog.Store(args.Id, reply.Value)
	// } else {

	// 	reply.Value = value.(string)

	// }
	// task <- args.Id
}

//	func (kv *KVServer) TaskArrager() {
//		for {
//			task := <-kv.taskchan
//			<-task
//			<-task
//		}
//	}
func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	DPrintf("Put request :%v \n at time: %s", args, kv.serverticker.Format(time.RFC3339Nano))
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if !args.Logmode {
		kv.log.Delete(args.Id)
		return
	}
	// task := make(chan int64)
	// kv.taskchan <- task
	// task <- args.Id

	if _, ok := kv.log.Load(args.Id); !ok {
		kv.kvsaver.Store(args.Key, args.Value)
		kv.log.Store(args.Id, args.Value)
	}
	// task <- args.Id
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	DPrintf("Append request :%v \n at time: %s", args, kv.serverticker.Format(time.RFC3339Nano))
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if !args.Logmode {
		kv.log.Delete(args.Id)
		return
	}
	// task := make(chan int64)
	// kv.taskchan <- task
	// task <- args.Id
	ov, ok := kv.log.Load(args.Id)
	if ok {
		reply.Value = ov.(string)
		return
	}
	value, ok := kv.kvsaver.Load(args.Key)
	if !ok {
		value = ""
	}
	reply.Value = value.(string)
	kv.kvsaver.Store(args.Key, value.(string)+args.Value)
	kv.log.Store(args.Id, reply.Value)

	// task <- args.Id
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	// You may need initialization code here.
	// 初始化
	kv.mu = sync.Mutex{}
	kv.log = sync.Map{}
	//kv.getlog = sync.Map{}
	kv.kvsaver = sync.Map{}
	// kv.taskchan = make(chan chan int64)
	//debug
	kv.serverticker = time.Now()
	// go kv.TaskArrager()
	return kv
}
