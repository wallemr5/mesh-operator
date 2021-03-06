package constant

var ZkServers = []string{
	// local
	//"10.12.210.70:2181",
	// tc-dev
	//"10.248.224.37:2181",
	// dev-sym
	"127.0.0.1:2181",
	// dev - dsf
	//"10.248.224.25:2181",

}

var (
	ApplicationLabel      = "application"
	AppCodeLabel          = "app_code"
	ProjectCodeLabel      = "pro_code"
	DubboPortName         = "dubbo-http"
	DubboProtocol         = "HTTP"
	MosnPort              = "20882"
	InstanceLabelZoneName = "zone"
	SourceLabelZoneName   = "sym-zone"
	ZoneValue             = "gz01"
	DefaultConfigName     = "DEFAULT_SERVICE"
)

var (
	PromHttpPort = ":8315"
)
