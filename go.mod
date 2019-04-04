module github.com/develar/app-builder

require (
	github.com/aclements/go-rabin v0.0.0-20170911142644-d0b643ea1a4c
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc // indirect
	github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf // indirect
	github.com/apex/log v1.1.0
	github.com/aws/aws-sdk-go v1.19.9
	github.com/biessek/golang-ico v0.0.0-20180326222316-d348d9ea4670
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/develar/errors v0.8.2
	github.com/develar/go-fs-util v2.0.1-0.20181113101504-f6630ccc0e93+incompatible
	github.com/develar/go-pkcs12 v0.0.0-20181115143544-54baa4f32c6a
	github.com/disintegration/imaging v1.6.0
	github.com/dustin/go-humanize v1.0.0
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/json-iterator/go v1.1.6
	github.com/jsummers/gobmp v0.0.0-20151104160322-e2ba15ffa76e // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mcuadros/go-version v0.0.0-20190308113854-92cdf37c5b75
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/oxtoacart/bpool v0.0.0-20190227141107-8c4636f812cc
	github.com/phayes/permbits v0.0.0-20190108233746-1efae4548023
	github.com/pkg/errors v0.8.1 // indirect
	github.com/pkg/xattr v0.4.1
	github.com/segmentio/ksuid v1.0.2
	github.com/stretchr/testify v1.3.0 // indirect
	github.com/zieckey/goini v0.0.0-20180118150432-0da17d361d26
	golang.org/x/image v0.0.0-20190321063152-3fc05d484e9f // indirect
	golang.org/x/net v0.0.0-20190403144856-b630fd6fe46b // indirect
	golang.org/x/sys v0.0.0-20190403152447-81d4e9dc473e // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

//replace github.com/develar/go-pkcs12 => ../go-pkcs12
