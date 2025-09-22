package dynamicProxyController

type Operation string

const (
	OperationBind   Operation = "bnd"
	OperationUnbind Operation = "ubd"
)

type Mapping struct {
	Id         int64     `json:"id"`
	Operation  Operation `json:"o"`
	Name       string    `json:"n"`
	ShareToken string    `json:"st"`
}
