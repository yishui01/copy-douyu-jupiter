package constant

type ServiceKind uint8

const (
	ServiceUnknown ServiceKind = iota
	ServiceProvider
	ServiceGovernor
	ServiceConsumer
)

var serviceKinds = make(map[ServiceKind]string)

func init() {
	serviceKinds[ServiceUnknown] = "unknown"
	serviceKinds[ServiceProvider] = "providers"
	serviceKinds[ServiceGovernor] = "governors"
	serviceKinds[ServiceConsumer] = "consumers"
}

func (sk ServiceKind) String() string {
	if s, ok := serviceKinds[sk]; ok {
		return s
	}
	return "unknown"
}
