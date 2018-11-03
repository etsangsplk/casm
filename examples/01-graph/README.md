# Stack Level 1:  Expander Graph

## Context

The graph layer connects `casm.Host`s into a mesh network over which
messages can be efficiently broadcast.  Note that this broadcast network is
**unreliable**; packets may be dropped or arrive out of order.

The CASM expander-graph model has a few noteworthy properties:

1. Each `graph.Vertex` maintains `k` connections to other vertices, where `k=5` by default.

2. Connections are random, and periodically shuffled to maintain random conectivity properties.

To avoid loops, vertices implement filters, which drop duplicate packets.  Two
filter types are commonly encountered:

1. **State Filters** treat the stream of broadcast messages as updates to a global state and drop packets that are "older" than thier most recently-seen counterparts.  This useful for e.g. discovering peers or services on a network, and represents the use-case for which CASM's graph model was designed.
2. **Stable Bloom Filters** treat the stream of broadcast messages as an unbounded stream of data.  Duplicates are removed by means of a [stable Bloom filter](http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.85.1569&rep=rep1&type=pdf) (SBF).  **N.B:** contrary to classical Bloom filters, SBFs introduce the possibility of **both** false-positives **and** false negatives.  In addition to dropping packets or delivering packets out-of-order, vertices using SBFs may deliver **duplicate** packets.  In practice, duplication should occur relatively locally in the graph, as each node will use a random seed for its SBF and is therefore unlikely to misclassify the same packets as its peers.

## Demonstration

In this example, we create a small graph, broadcast some messages, and show that
all vertices received all messages.
