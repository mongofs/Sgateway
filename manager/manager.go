package manager

import "Sgateway/proxy"

type Manager interface {
	// 添加proxy ，一般不会涉及到非常高频的操作，添加过后生效就行
	Add(proto proxy.Protocol, proxy *proxy.Proxy)error

	// 删除url ，如果不存在就报错出去
	Del(identification string)error

	// 查询某个种类的
	Select(proto proxy.Protocol)
}
