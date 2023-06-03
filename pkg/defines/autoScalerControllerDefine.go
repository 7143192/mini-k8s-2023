package defines

const AutoScalerControllerPrefix = "AutoScalerController"

type AutoScalerController struct {
	AutoScalers []*EtcdAutoScaler `yaml:"autoScalers" json:"autoScalers"`
}
