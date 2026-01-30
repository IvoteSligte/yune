using Go = import "go-capnp/std/go.capnp";
@0xeaa1f06c3c8d5689; # file ID
$Go.package("pb");
$Go.import("yune/pb");

struct Value {
  struct Type {
    union {
      type   @0 :Void;
      int    @1 :Void;
      float  @2 :Void;
      bool   @3 :Void;
      string @4 :Void;
      nil    @5 :Void;
      fn :group {
        argument @6 :Type;
	return   @7 :Type;
      }
      tuple  @8 :List(Type);
      list   @9 :Type;
      struc :group { # struct type
        name @10 :Text;
      }
    }
  }
  union {
    nothing @0 :Void;
    type @1 :Type;
  }
}
