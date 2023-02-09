module github.com/bmc-toolbox/actor

go 1.12

require (
	github.com/GeertJohan/go.rice v1.0.2
	github.com/aws/aws-sdk-go v1.37.5
	github.com/bmc-toolbox/bmclib v0.4.15
	github.com/bmc-toolbox/gin-go-metrics v0.0.2
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/gin-gonic/gin v1.7.7
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/ugorji/go v1.2.4 // indirect
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
