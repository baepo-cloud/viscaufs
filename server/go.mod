module github.com/baepo-cloud/viscaufs-server

go 1.24.2

require (
	github.com/alphadose/haxmap v1.4.1
	github.com/amacneil/dbmate/v2 v2.27.0
	github.com/baepo-cloud/viscaufs-common v0.0.0-00010101000000-000000000000
	github.com/google/go-containerregistry v0.20.3
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.2
	github.com/joho/godotenv v1.5.1
	github.com/nrednav/cuid2 v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/plar/go-adaptive-radix-tree/v2 v2.0.3
	github.com/stretchr/testify v1.10.0
	go.uber.org/fx v1.23.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.26.0
)

require (
	github.com/containerd/stargz-snapshotter/estargz v0.16.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v27.5.0+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-sqlite3 v1.14.28 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vbatts/tar-split v0.11.6 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250428153025-10db94c68c34 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/baepo-cloud/viscaufs-common => ../common
