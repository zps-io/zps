module github.com/fezz-io/zps

go 1.13

require (
	cloud.google.com/go v0.75.0
	github.com/DataDog/zstd v1.4.1 // indirect
	github.com/Sereal/Sereal v0.0.0-20190618215532-0b8ac451a863 // indirect
	github.com/asdine/storm v2.1.2+incompatible
	github.com/aws/aws-sdk-go v1.24.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/chuckpreslar/emission v0.0.0-20170206194824-a7ddd980baf9
	github.com/davecgh/go-spew v1.1.1
	github.com/dsnet/compress v0.0.1
	github.com/fezz-io/sat v0.0.0-20190412034122-acaa8fa26246
	github.com/gernest/wow v0.1.0
	github.com/golang/snappy v0.0.1 // indirect
	github.com/hashicorp/hcl/v2 v2.3.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/lunixbochs/struc v0.0.0-20190916212049-a5c72983bc42
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/naegelejd/go-acl v0.0.0-20190510140445-686b8e62cbee
	github.com/nightlyone/lockfile v0.0.0-20180618180623-0ad87eef1443
	github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/segmentio/ksuid v1.0.2
	github.com/spf13/cobra v0.0.5
	github.com/zclconf/go-cty v1.3.1
	go.etcd.io/bbolt v1.3.3
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	gonum.org/v1/gonum v0.0.0-20190915125329-975d99cd20a9
	gopkg.in/resty.v1 v1.12.0
)

replace github.com/gernest/wow v0.1.0 => github.com/fezz-io/wow v0.1.1-0.20200606051511-4eedecafd068
