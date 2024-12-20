<b>A HTTP BASED distributed cache, supporting following requirements
- HTTP protocol for server-client communication
- P2P for data replication
- Sharding by key using consistent hashing algorithm
- TTL and LRU based eviction policies.
</b>
**INFO:

Cache eviction policies

We cant keep all data in memory indefinitely. 
1. LRU : doubly-linked list in standart library "containers/list"
2. TTL: evicted items based on tim rather than usage
3. FIFO: evicts oldest item in the cache based on the time it was added, very easy to implement, frequently used items 

Choosing the right eviction policy
LRU and TTL

- Adding TTL: a seperate goroutine with time.Ticker, periodically triggers evict function to check and remove expired entries.
 eviction during GET: no need to manage extra goroutine which leads to less complex code, reduces overhead.
 but also cons: delayed eviction. potential latency on GET

 Replication

 P2P: all nodes can both send ands receive updates(writes). Each node can communicate with any other node, distributnging load evenly across
 the network. This allows the systen to scale horizontally more efficiently as there is no single point of contention. LEADERLESS architecture.
 Flexbility is good , nodecan can join or leave the network.

 Pub/sub: using message broker to broadcaset updates to replicas, scalability depends on broker performance and architecture.

 Primary replica: Scalability is limited because leader node can become bottleneck as number of client increases.

 Distributed consensus protocols: Most known ones are Raft and Paxos. It is good for strong consistency and smaller clusters. 

 "We will be going with P2P"

 Sharding

 partition data across multiple nodes, ensuring scalability and performance
 benefits:
    - horizontal scaling: allows you to scale out by adding more nodes(shards) to your system. 
    - load distribution: by distributing data across multiple shards, helps balance the load.
 approaches for sharding
 1. range based: 
 2. hash based: a hash function is applied to the shard key to determine the shard. range queries are inefficient as they may span multiple shards.
 3. consistent hashing:  each node is responsible for keys in its range. nodes and keys are hashed to a circular space. more complex. Each cache server is assigned to
 a random position on the ring. moving clockwise

 "We will be going with consistent hashing
