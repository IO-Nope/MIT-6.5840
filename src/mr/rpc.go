package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"

	"6.824/mr"
)

// example to show how to declare the arguments
// and reply for an RPC.
type MapArgs struct {
	Routinenum int
}
type MapReply struct {
	Filename string
}
type ReduceArgs struct {
	Routinenum int
}
type ReduceReply struct {
	Isdone bool
}
type MaperrArgs struct {
	Filename string
}
type EmitMArgs struct {
	Intermediate []mr.KeyValue
}

type MaperrReply struct {
	Isdone bool
}
type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
