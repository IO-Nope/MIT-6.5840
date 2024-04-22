# MIT-6.5840 分布式系统

[课程链接](https://pdos.csail.mit.edu/6.824/index.html)
## 完成情况
 - [x] Lab1: MapReduce
 - [ ] Lab2A: Raft, leader election
 - [ ] Lab2B: Raft, log
 - [ ] Lab2C: Raft, persistence
 - [ ] Lab2D: Raft, log compaction
 - [ ] Lab3A: Fault-tolerant Key/Value Service without snapshots
 - [ ] Lab3B: Fault-tolerant Key/Value Service with snapshots
 - [ ] Lab4: Sharded Key/Value Service

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
```
 

```



