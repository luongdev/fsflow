package shared

//type ClientWorkflow struct {
//	FsClient *freeswitch.SocketClient
//}
//
//func NewClientWorkflow(fsClient *freeswitch.SocketClient) *ClientWorkflow {
//	return &ClientWorkflow{
//		FsClient: fsClient,
//	}
//}
//
//type ClientActivity struct {
//	FsClient *freeswitch.SocketClient
//}
//
//func NewClientActivity(fsClient *freeswitch.SocketClient) *ClientActivity {
//	return &ClientActivity{
//		FsClient: fsClient,
//	}
//}

type FreeswitchWorkflow interface {
	Handler() WorkflowFunc

	Name() string
}

type FreeswitchActivity interface {
	Handler() ActivityFunc

	Name() string
}
