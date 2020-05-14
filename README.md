# go2proto

Generate Protobuf messages from given go structs. No RPC, not gogo syntax, just pure Protobuf messages.

Forked from [github.com/anjmao/go2proto](https://github.com/anjmao/go2proto)

### Example

```sh
GO111MODULE=off go get -u github.com/emarcey/go2proto
go2proto -f ${PWD}/example/out -p github.com/emarcey/go2proto/example/in
```

### Configuration

* `-f`: directory of go files to convert to proto messages
* `-p`: target directory for output proto
* `filter`: if set, excludes all structs not containing this string
* `-c`: current proto, path of existing version of proto to use for diff
