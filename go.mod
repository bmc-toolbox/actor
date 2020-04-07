module github.com/bmc-toolbox/actor

go 1.12

require (
	github.com/GeertJohan/go.rice v1.0.0
	github.com/aws/aws-sdk-go v0.0.0-20190114232201-beaa15b1b227
	github.com/bmc-toolbox/bmclib v0.3.6
	github.com/bmc-toolbox/gin-go-metrics v0.0.1
	github.com/gin-gonic/gin v1.3.0
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.0-20181021141114-fe5e611709b0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.3.2
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422 // indirect
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
