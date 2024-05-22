package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 3D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 3D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}
type LogEntry struct {
	LogTerm int
	Command string
}
type Timer struct {
	time int
	mu   sync.Mutex
}
type State int

const (
	Follower State = iota
	Candidate
	Leader
)

type votedFor struct {
	vote int
	mu   sync.Mutex
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (3A, 3B, 3C).
	// 3A
	currentTerm int
	votedFor    votedFor
	log         []LogEntry
	beatTimer   Timer
	state       State
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (3A).
	term = rf.currentTerm
	isleader = (rf.state == Leader)
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).
	Term        int
	CandidateId int
	// LastLogIndex int
	// LastLogTerm  int

}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).
	Term        int
	VoteGranted bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3A, 3B).

	if args.Term < rf.currentTerm {
		DPrintfA("Server %d refuse to vote for %d because it's term", rf.me, args.CandidateId)
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		return
	}
	rf.mu.Lock()
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = Follower
		rf.votedFor.mu.Lock()
		rf.votedFor.vote = -1
		rf.votedFor.vote = args.CandidateId
		reply.VoteGranted = true
		reply.Term = rf.currentTerm
		DPrintfA("Server %d vote for %d at election term %d because out of term", rf.me, args.CandidateId, rf.currentTerm)
		rf.votedFor.mu.Unlock()
		rf.mu.Unlock()
		return
	}
	rf.mu.Unlock()
	rf.votedFor.mu.Lock()
	defer rf.votedFor.mu.Unlock()
	if rf.votedFor.vote == -1 && rf.state != Leader {
		DPrintfA("Server %d vote for %d at election term %d", rf.me, args.CandidateId, rf.currentTerm)
		rf.votedFor.vote = args.CandidateId
		reply.VoteGranted = true
		reply.Term = rf.currentTerm
		return
	}
	DPrintfA("Server %d refuse to vote for %d", rf.me, args.CandidateId)
	reply.VoteGranted = false
	reply.Term = rf.currentTerm
}

// AppendEntries RPC arguments structure.
type AppendEntriesArgs struct {
	Term     int
	LeaderId int
	// PrevLogIndex int
	// PrevLogTerm int
	Entries []LogEntry
	// LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	if args.Term < rf.currentTerm {
		reply.Success = false
		reply.Term = rf.currentTerm
		return
	}
	if args.Term > rf.currentTerm {
		rf.mu.Lock()
		rf.currentTerm = args.Term
		rf.state = Follower
		rf.mu.Unlock()
		reply.Success = true
		reply.Term = rf.currentTerm
		DPrintfA("server %d change their term because leader %d at term %d", rf.me, args.LeaderId, args.Term)
	}
	if rf.votedFor.vote != args.LeaderId {
		rf.votedFor.mu.Lock()
		rf.votedFor.vote = args.LeaderId
		rf.votedFor.mu.Unlock()
	}
	rf.beatTimer.mu.Lock()
	rf.beatTimer.time = 2
	rf.beatTimer.mu.Unlock()
	//3B
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}
func (rf *Raft) sendHreatBeats() {
	for rf.state == Leader {
		for i := range rf.peers {
			if i == rf.me {
				continue
			}
			go func(i int) {
				args := AppendEntriesArgs{
					Term:     rf.currentTerm,
					LeaderId: rf.me,
					Entries:  nil,
				}
				reply := AppendEntriesReply{
					Term:    0,
					Success: false,
				}
				rf.sendAppendEntries(i, &args, &reply)
			}(i)
		}
		time.Sleep(time.Duration(300) * time.Millisecond)
	}
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (3B).

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}
func (rf *Raft) election(ctx context.Context, temchan chan bool) {
	if rf.state != Candidate {
		DPrintfA("election stop because isn't Candidate")
		return
	}
	rf.mu.Lock()
	rf.currentTerm++
	rf.votedFor.mu.Lock()
	rf.votedFor.vote = rf.me
	rf.votedFor.mu.Unlock()
	rf.state = Candidate
	rf.mu.Unlock()
	votes := 1
	voted := 1
	endchan := make(chan bool, 1)
	tempmu := sync.Mutex{}
	tempcond := sync.NewCond(&tempmu)
	for i := range rf.peers {
		select {
		case <-ctx.Done():
			DPrintfA("Server %d time out", rf.me)
			return
		case <-endchan:
			return
		default:
		}
		if i == rf.me {
			continue
		}
		go func(i int) {
			args := RequestVoteArgs{
				Term:        rf.currentTerm,
				CandidateId: rf.me,
			}
			var reply = RequestVoteReply{
				Term:        0,
				VoteGranted: false,
			}
			rf.sendRequestVote(i, &args, &reply)
			if voted >= len(rf.peers) {
				return
			}
			rf.mu.Lock()
			if reply.Term > rf.currentTerm {
				DPrintfA("Server %d out of term", rf.me)
				rf.currentTerm = reply.Term
				rf.state = Follower
				rf.votedFor.mu.Lock()
				rf.votedFor.vote = -1
				rf.votedFor.mu.Unlock()
				rf.mu.Unlock()
				tempmu.Lock()
				voted = len(rf.peers) + 1
				tempcond.Signal()
				tempmu.Unlock()
				return
			}
			rf.mu.Unlock()
			tempmu.Lock()
			defer tempmu.Unlock()
			voted++
			if reply.VoteGranted {
				votes++
			}
			tempcond.Signal()

		}(i)
	}
	tempmu.Lock()
	for votes < len(rf.peers)/2+1 {
		if voted >= len(rf.peers) {
			DPrintfA("Server %d failed at becoming a leader", rf.me)
			if rf.state == Candidate {
				rf.mu.Lock()
				rf.state = Follower
				rf.mu.Unlock()
			}
			tempmu.Unlock()
			return
		}
		tempcond.Wait()
	}
	tempmu.Unlock()
	rf.mu.Lock()
	select {
	case <-ctx.Done():
		DPrintfA("Server %d time out", rf.me)
		rf.state = Follower
		rf.mu.Unlock()
		endchan <- true
		return
	default:
	}
	if rf.state != Candidate {
		rf.mu.Unlock()
		DPrintfA("Server %d failed beacause it is a follower!", rf.me)
		endchan <- true
		return
	}
	DPrintfA("%d get %d votes of %d peers and voted %d", rf.me, votes, len(rf.peers), voted)
	DPrintfA("Server %d become a leader at term %d", rf.me, rf.currentTerm)
	rf.state = Leader
	endchan <- true
	temchan <- true
	go rf.sendHreatBeats()
	rf.mu.Unlock()

}
func (rf *Raft) ticker() {
	for !rf.killed() {

		// Your code here (3A)
		// Check if a leader election should be started.
		if rf.beatTimer.time <= 0 && rf.state != Leader {
			rf.mu.Lock()
			rf.state = Candidate
			rf.mu.Unlock()
			DPrintfA("Server %d start election at Term %d", rf.me, rf.currentTerm+1)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			suchan := make(chan bool)
			fachan := make(chan bool)
			go rf.election(ctx, suchan)
			ms := 30 + (rand.Int63() % 170)
			go func() {
				time.Sleep(time.Duration(ms) * time.Millisecond)
				fachan <- true
			}()
			select {
			case <-fachan:
				DPrintfA("Server %d timeout", rf.me)
				cancel()

			case <-suchan:
				DPrintfA("Server %d finish election", rf.me)
				cancel()
			}

		}
		rf.beatTimer.mu.Lock()
		rf.beatTimer.time--
		rf.beatTimer.mu.Unlock()
		// pause for a random amount of time between 50 and 350
		// milliseconds.
		ms := 150 + (rand.Int63() % 200)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3A, 3B, 3C).
	//3A
	rf.mu = sync.Mutex{}
	rf.currentTerm = 0
	rf.votedFor = votedFor{
		vote: -1,
		mu:   sync.Mutex{},
	}
	rf.log = make([]LogEntry, 0)
	rf.beatTimer = Timer{
		time: 0,
		mu:   sync.Mutex{},
	}
	rf.state = Follower
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}
