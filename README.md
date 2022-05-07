# CRDT [![Go Report Card](https://goreportcard.com/badge/github.com/bjornaer/crdt)](https://goreportcard.com/report/github.com/bjornaer/crdt) ![tests](https://github.com/bjornaer/crdt/actions/workflows/push.yaml/badge.svg)[![HitCount](https://hits.dwyl.com/bjornaer/crdt.svg?style=flat-square)](http://hits.dwyl.com/bjornaer/crdt)

### Introduction
Conflict-Free Replicated Data Types (CRDTs) are data structures that power real-time collaborative applications in
distributed systems. CRDTs can be replicated across systems, they can be updated independently and concurrently
without coordination between the replicas, and it is always mathematically possible to resolve inconsistencies that
might result.

In this codebase we implement a LWW-Element-Set based Graph.
### Last-Write-Wins-Element-Set
LWW-Element-Set is similar to 
[2P-Set](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type#2P-Set_(Two-Phase_Set)) 
in that it consists of an "add set" and a "remove set",
with a timestamp for each element. 
Elements are added to an LWW-Element-Set by inserting the element into the add set, with a timestamp.
Elements are removed from the LWW-Element-Set by being added to the remove set, again with a timestamp.
An element is a member of the LWW-Element-Set if it is in the add set, and either not in the remove set,
or in the remove set but with an earlier timestamp than the latest timestamp in the add set. Merging two replicas of the
LWW-Element-Set consists of taking the union of the add sets and the union of the remove sets.
When timestamps are equal, the "bias" of the LWW-Element-Set comes into play.
A LWW-Element-Set can be biased towards adds or removals.
An advantage of LWW-Element-Set is that it allows an element to be reinserted after having been removed.

### Package

This package implements a `CRDT` interface that enables use of the `LWW-Graph` structure using a `LWW-Element-Set` to represent its set of vertices while for its edges it uses a `mapping of a vertex to a LWW-Element-Set` representing all edges of said vertex.

The package also exposes the option to simply use a `LWW-Element-Set`.

As stated in the previous section the `LWW-Element-Set` contains both `Additions` and `Removals` sets, 
to where we monotonically add graph elements marked to be added or removed. 
To make sure we keep things monotonically, the LWW-Element-Set is built using a helper structure we call `Time Map`.
The `Time Map` is an abstraction of a Go Map in which we only allow the operation of adding elements,
and map those elements to a timestamp in the moment of the addition.
Thus, only allowing items to be added to both the `Adittions` and `Removals` set.

### Roadmap

I intend on adding support for a more varied options of backends such as Redis
I also intend on keep adding more types of CRDTs to be used besides the `LWW-Set` and `LWW-Graph`

---
**NOTE**

To read documentation on the public API [can be found here](https://pkg.go.dev/github.com/bjornaer/crdt)
---

### Run tests

To run tests

```go
go test ./...
```

### Bibliography

- [A comprehensive study of Convergent and Commutative Replicated Data Types](https://hal.inria.fr/file/index/docid/555588/filename/techreport.pdf)
- [Consistency without consensus in production systems by Peter Bourgon](https://www.youtube.com/watch?v=em9zLzM8O7c)
- [Roshi: a CRDT system for timestamped events](https://developers.soundcloud.com/blog/roshi-a-crdt-system-for-timestamped-events)
- [CRDT notes by Paul Frazee](https://github.com/pfrazee/crdt_notes)
- [Wikipedia page on CRDT](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type)