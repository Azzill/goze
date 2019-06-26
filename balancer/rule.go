package balancer

type LoadBalanceRule string

const (
	_                      LoadBalanceRule = ""
	WeightedRoundRobinRule                 = "wwr"
	AddressHash                            = "addr_hash"
)
