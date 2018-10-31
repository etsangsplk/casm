# capnp compile -I$GOPATH/src/zombiezen.com/go/capnproto2/std -ogo api/graph/message.capnp
using Go = import "/go.capnp";
@0xf9eb5f58886df78d;
$Go.package("graph");
$Go.import("github.com/lthibault/casm/api/graph");

struct Message $Go.doc("Message for broadcast over the graph") {
    id @0 :UInt64;
    seq @1 :UInt64;
    pub :union {
        none @2 :Void;
        topic @3 :Data;
    }
    body @4 :Data;
}
