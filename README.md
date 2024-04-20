# MIT-6.5840 分布式系统

[课程链接](https://www.example.com)
## 完成情况
 - [ ] Lab1: MapReduce
 - [ ] Lab2A: Raft, leader election
 - [ ] Lab2B: Raft, log
 - [ ] Lab2C: Raft, persistence
 - [ ] Lab2D: Raft, log compaction
 - [ ] Lab3A: Fault-tolerant Key/Value Service without snapshots
 - [ ] Lab3B: Fault-tolerant Key/Value Service with snapshots
 - [ ] Lab4: Sharded Key/Value Service

---


## lab1: MapReduce

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
            rules规定reduce任务需要分成nReduce个数并将中间结果存储在mr-out-X-Y.txt中要一行一行，其中X是Mapid Y是ihash得到的Reduceid
- [ ] 制作相应结构
    - MapJabs 
    ```
    ```
    - Coordinator
    ```
    ```
    - RPCS
    ```
    ```
---
```
 

```



