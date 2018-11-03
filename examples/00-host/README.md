# Stack Level 0: Raw Hosts

Hosts are the lowest-level abstraction in CASM, and are represented by the `casm.Host` interface.  A host is a logical peer in the network, and will usually
correspond to a physical machine.

Hosts are responsible for establishing network connections between each other,
and for multiplexing `casm.Stream`s on top of these connections.

This example shows how to establish a bidirectional stream between two `casm.Host` instances.  We start by initializing two hosts that listen on the loopback
interface.  We then register a stream handler echos back any data received.
Finally, the other peer transmits data and prints out what it receives.
