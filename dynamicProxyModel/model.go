package dynamicProxyModel

type Operation string

const (
	OperationBind   Operation = "bi"
	OperationUnbind Operation = "un"
)

const MappingUpdate = "mup"

type Mapping struct {
	Operation  Operation `json:"o"`
	Name       string    `json:"n"`
	Version    int64     `json:"v"`
	ShareToken string    `json:"st"`
}
