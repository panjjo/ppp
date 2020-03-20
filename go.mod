module github.com/panjjo/ppp/v1.1.0

go 1.13

require (
	github.com/golang/protobuf v1.3.5
	github.com/panjjo/ppp v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/viper v1.6.2
	gogs.yunss.com/go/k8s v0.0.0-20191115021454-06a7ca2c753e
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	google.golang.org/grpc v1.28.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
)

replace github.com/panjjo/ppp => ../ppp
