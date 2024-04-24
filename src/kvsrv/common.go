package kvsrv

// Put or Append
type PutAppendArgs struct {
	Key   string
	Value string
	// You'll have to add definitions here.
	// 独特的id使得每个请求都是唯一的
	Id      int64
	Logmode bool
	// Field names must start with capital letters,
	// otherwise RPC will break.
}

type PutAppendReply struct {
	Value string
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
	//独特的id使得每个请求都是唯一的
	Id      int64
	Logmode bool
}

type GetReply struct {
	Value string
}
