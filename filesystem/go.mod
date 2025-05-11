module github.com/baepo-cloud/viscaufs-fs

go 1.24.2

require (
	github.com/alphadose/haxmap v1.4.1
	github.com/baepo-cloud/viscaufs/common v0.0.0-00010101000000-000000000000
	github.com/hanwen/go-fuse/v2 v2.7.2
	google.golang.org/grpc v1.72.0
)

require (
	github.com/alexisvisco/go-adaptive-radix-tree/v2 v2.0.0-20250510163150-cd486f626aff // indirect
	golang.org/x/exp v0.0.0-20221031165847-c99f073a8326 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/baepo-cloud/viscaufs/common => ../common
