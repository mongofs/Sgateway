package proxy

import "net/http"

type Protocol uint8

const (
	ProtocolHttp  Protocol = iota + 1 // http
	protocolHTTPS                     // https

)

type Proxy interface {
	// Reverse 反转代理， 将请求代理到Reverse参数url中
	Reverse(request *http.Request, write http.ResponseWriter) error

	// Protocol 获取当前代理类型
	Protocol() string

	// loadConn 当前负载的请求数量
	LoadConn() int32

	// Weight 获取到当前的weight 负载
	Weight () int

	// Target 获取到目标地址的url
	Target () string
}

type Action struct {

	// proxy 代理，用户访问是就需要设置proxy
	proxy Proxy

	// limit 限流，用户进行访问时候，如果出现超载就需要进行限流
	limit Limit

	AddTime int64

	// 当前代理状态
	Status uint8
}
