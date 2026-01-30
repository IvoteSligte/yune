# Yune

## Dependencies

clang++
clang-format

Cap'n Proto compiler (`capnp`)
Cap'n Proto Go plugin (`go install capnproto.org/go/capnp/v3/capnpc-go@latest`)
Cap'n Proto development files (C++ headers)

Note that the Cap'n Proto development files are currently also needed to use the compiler as they are dynamically linked. They will likely be included in the future, if Cap'n Proto remains as the serialization method of choice.
