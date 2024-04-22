package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}
type JobRequestArgs struct {
	AskJob bool
}
type JobDoneArgs struct {
	Job      *Job
	Workerid int
}
type JobDoneReply struct {
	IsReceive bool
}
type JobRequestReply struct {
	Job      *Job
	Workerid int
}
type HreatBeatArgs struct {
	WorkerId int
}
type HreatBeatReply struct {
	IsReceive bool
}
type JobCheckArgs struct {
	Jobid    int
	Workerid int
}
type JobCheckReply struct {
	IsAllow bool
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
