module github.com/develar/app-builder

require (
	github.com/aclements/go-rabin v0.0.0-20170911142644-d0b643ea1a4c
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/alessio/shellescape v0.0.0-20190409004728-b115ca0f9053 // indirect
	github.com/aws/aws-sdk-go v1.38.68
	github.com/biessek/golang-ico v0.0.0-20180326222316-d348d9ea4670
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/develar/errors v0.9.0
	github.com/develar/go-fs-util v0.0.0-20190620175131-69a2d4542206
	github.com/develar/go-pkcs12 v0.0.0-20181115143544-54baa4f32c6a
	github.com/disintegration/imaging v1.6.2
	github.com/dustin/go-humanize v1.0.0
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/json-iterator/go v1.1.11
	github.com/jsummers/gobmp v0.0.0-20151104160322-e2ba15ffa76e // indirect
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.13
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/phayes/permbits v0.0.0-20190612203442-39d7c581d2ee
	github.com/pkg/errors v0.9.1
	github.com/pkg/xattr v0.4.3
	github.com/segmentio/ksuid v1.0.3
	github.com/zieckey/goini v0.0.0-20180118150432-0da17d361d26
	go.uber.org/atomic v1.8.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.17.0
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	gopkg.in/alessio/shellescape.v1 v1.0.0-20170105083845-52074bc9df61
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
)

//replace github.com/develar/go-pkcs12 => ../go-pkcs12

go 1.13
