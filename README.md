# MIT-6.5840 分布式系统

[课程链接](https://pdos.csail.mit.edu/6.824/index.html)
## 完成情况
 - [x] Lab1: MapReduce
 - [x] Lab2: Key/Value Server
 - [x] Lab3A: Raft, leader election
 - [ ] Lab3B: Raft, log
 - [ ] Lab3C: Raft, persistence
 - [ ] Lab3D: Raft, log compaction
 - [ ] Lab4A: Fault-tolerant Key/Value Service without snapshots
 - [ ] Lab4B: Fault-tolerant Key/Value Service with snapshots
 - [ ] Lab5: Sharded Key/Value Service

---


## lab1: MapReduce 
完成状态: 
- [x] 已完成
### 思路:
- [x] 了解mrsequential.go的逻辑和lab1代码的逻辑
- [x] 了解coordinator和worker之间的RPC通信
    - coordinator负责调度任务和打包任务 
    - worker通过发送RPC请求来申请任务
    - RPCcall通过call函数中args与replay与查找到的string指定函数交互并修改args与replay值来实现交互
- [x] 规划功能
    - coordinator负责
        - 接收Files的路径作为参数,由此制作Map任务.
        - 任务被放在一个jobchannel中等待worker发送RPC来申请
        - 每个任务要有相应的信息
        - 在有Map任务完成后制作Reduce任务
        - 维护任务状态
            - 切换任务形态
            - 任务失败后使其能够再度被发布
    - worker负责
        - 发送RPC向coordinator请求任务
        - 收到任务后按照任务种类分配任务
            - Map任务:
            <-filename中用mrsequential.go中的代码将其分割成kv对 包含着任务id Emit出去给coordinator
            - Reduce任务:
            rules规定reduce任务需要分成nReduce个数并将中间结果存储在mr-out-X-Y.txt中要一行一行，其中X是Mapid Y是ihash(key)%nReduce得到的Reduceid
        - 在结束任务后继续请求,请求不到任务的时候sleep一会再请求,直到coordinator关闭(ps.worker申请不到RPC自动关闭)或者接到DoneJob
- [x] 搞清楚测试脚本的逻辑,需要从中得到哪里发生错误
### 测试代码:
```
    cd src/main
    bash test-mr-sh
```
### Hints/遭遇
- 不仅是RPC中定义通讯的结构本身成员首字母需要大写,其成员所包括的成员也要大写,不然在传递过程中会被隐式初始化
- 测试test-mr.sh是以src/main/temp(它临时生成的文件夹)为基本路径,注意和自己在src/main下直接跑访问相对路径的情况不一样
- 避免在学业繁重的时候写代码,不然被测试脚本气晕的晚上根本学不进去！

### Todo/后记
- [ ] 优化coordinator里一堆waitgroup和mutex的问题
- [ ] 完善workerid的问题,现在的workerid更像是coordinator的workjobid
- [ ] 通过Lab1的No-Credit考验
---
## Lab2: Key/Value Server
> 注: 这个Lab是2024新加的

完成状态: 
- [x] 已完成
### 思路:
- 项目架构
- 测试思路
- 笔记和思路51补吧,没精力了
- 
### 测试代码:
```
    cd src/kvsrv
    go test
```
### Hints/遭遇
- 我是大傻逼！我是大傻逼！我是大傻逼！我是大傻逼！我是大傻逼！我是大傻逼！我是大傻逼！我是大傻逼！
- 弄清楚需求!搞清楚错在哪!看一看最基本的地方！
- 这个项目里注释掉了一堆用来实现"并行线性化"的东西,它在这个Lab里是完全不需要的！
--- 
## Lab3A: Raft, leader election
> ps. 有个大**五一爆玩啥都没做

完成状态: 
- [x] 已完成
### 框架逻辑
> 以下是Raft paper里关于每个Raft的State描述
```
State：

所有服务器上的持久化状态，在回复RPC之前更新持久化存储
currentTerm：服务器知道的最近任期，当服务器启动时初始化为0，单调递增
votedFor：当前任期中，该服务器给投过票的candidateId，如果没有则为null
log[]：日志条目；每一条包含了状态机指令以及该条目被leader收到时的任期号

2. 所有服务器上的易失性状态
commitIndex：已知被提交的最高日志条目索引号，一开始是0，单调递增
lastApplied：应用到状态机的最高日志条目索引号，一开始为0，单调递增
```
- [ ] 
- [ ] 
- [ ] 
- [ ] 
- [ ] 
### 功能要求
- [x] 实现 Raft Leader选举 (单个领导者)
    - 在没有失败时让 Leader 继续担任
    - 如果旧 Leader 失败或旧 Leader 的数据包丢失，则新 Leader 可以接管
    - 任何状态peer收到任何比其任期高的信息都需要转变成follower
> 以下是Raft paper里关于选举的实现描述
```
RequestVote RPC，candidate发起来收集投票：

1. 参数 (Args)
term：candidate的任期号
candidateId：发起投票的candidate的ID
lastLogIndex：candidate的最高日志条目索引
lastLogTerm：candidate的最高日志条目的任期号

2. 结果 (Replys)
term：服务器的当前任期号，让candidate更新自己
voteGranted：如果是true，意味着candidate收到了选票

3. 接收者实现 (Call)
如果参数中的term < 接收者的currentTerm，返回false
如果服务器中的votedFor是null或者参数中的candidateId，而且candidate的日志至少和接收者的日志一样新(up-to-date)，获得选票
```
> 以下是lab要求
```
您无法轻松地直接运行 Raft 实例;相反，你应该通过测试运行它，即go test -run 3A。
按照论文的Figure 2。此时您只需要关心发送和接收 RequestVote RPC、与选举相关的服务器规则以及相关的情况，
将 Leader 选举的Figure 2 状态添加到raft.go上的 Raft结构中。您还需要定义一个结构体来保存有关每个日志条目的信息。
填写RequestVoteArgs和RequestVoteReply结构。修改Make()以创建一个后台 go 协程，该协程将在一段时间未从其他 peers 那里听到请求投票 RPC 时，发送RequestVote RPC 来定期启动 Leader 选举。这样，如果已经有一个 Leader，或者其成为 Leader，其他 peers 就会知道谁是Leader。实现RequestVote() RPC 函数，以便服务器投票给别人。
为了实现心跳检测，请提前定义AppendEntries RPC 结构（尽管您可能还不需要所有参数），并让 Leader 定期发送它们。AppendEntries RPC 函数需要重置选举超时时间，以便其他服务器已当选时，该不会以 Leader 的身份继续运行。
确保不同 Peers 不会在同一时间选举超时，否则所有 Peers 将只为自己投票，没有人会成为 Leader。
测试要求 Leader 发送心跳检测 RPC 的频率不超过 10 次/ 秒。
测试要求您的 Raft 在旧 Leader 失败后五秒内选出新 Leader（如果大多数同行仍然可以沟通）。但是，请记住，在发生分裂投票的情况下（如果数据包丢失或候选人不幸地选择相同的随机回票时间，则可能发生），领导人选举可能需要多轮投票。您必须选择足够短的选举超时（心跳间隔也是如此），确保即使选举需要多次轮断，也能在五秒内完成。
论文第 5.2 节提到选举超时应该在 150 到 300 毫秒范围内。只有当 Leader 发送一次心跳包的远小于 150 毫秒，这种范围才有意义。由于测试将您发送心跳包的频率限制在 10 次/秒内（译者注：也就是大于 100 毫秒），因此您必须使用比论文 150 到 300 毫秒更大的选举超时时间，但请不要太大，因为那可能导致无法在 5 秒内选出 Leader。
你可能会发现 Go 的 rand 很有用。
您将需要定期执行某些操作，或在一段时间后做些什么。最简单的方法是新起一个协程，在协程的循环中调用time.Sleep()。不要使用time.Timer或time.Ticker，这两个并不好用，容易出错（译者注：见仁见智）。
阅读有关锁和并发结构的建议。
如果代码在通过测试时遇到问题，请再次阅读论文的 Figure 2 ；Leader 选举的逻辑分布在Figure 2 的多个部分。
别忘了实现GetState()。
测试调用您的 Raft 的rf.Kill()时，您可以先调用rf.killed()再决定是否rf.Kill。您可能希望在所有循环中执行此功能，以避免已经死亡的 Raft 实例打印令人困惑的信息。
调试代码的一个好方法，就是在 Peer 发送或收到消息时打印自己的状态，并在测试时运行go test -run 2A > out，将日志收集到文件中。然后，通过研究 out 文件，可以确定实现中不正确的地方。您可能会喜欢用util. go中的Dprintf函数来调试，其可以在不同情况下打开和关闭日志。
Go RPC 仅发送以大写字母为首的结构体字段（译者注：可导出的字段）。子结构体还必须具有大写字段名称（例如数组中的日志记录字段）。labgob包会警告您这一点，不要忽略警告。
用go test -race测试你的代码，并修复它报告的任何问题。
```
### 思路:

