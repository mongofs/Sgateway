package proxy

import (
	"Sgateway/pkg/errors"
	"hash/crc32"
	"math/rand"
)

type LoadBalancing int

const (

	// 轮询
	Random LoadBalancing = iota

	// 最少连接数轮询
	Round

	// 权重
	WeightServer

	// 路径的hash
	UrlAddrHash
)

type (
	loadBalancer interface {
		add(items []Proxy)
		next(targetUrl string) Proxy
	}

	randomLoadBalance struct {
		nextLoopIndex int
		proxySet      []Proxy
	}

	roundLoadBalancer struct {
		nextLoopIndex int
		proxySet      []Proxy
		size          int
	}

	weightNode struct {
		proxy        Proxy
		sourceWeight int
		curWeight    int
	}

	weightServerLoadBalancer struct {
		proxySet []*weightNode
		curIdx   int // 当前的proxy
		curTurns int // 当前分配的轮次
		size     int // 当前分配的尺寸
	}

	urlAddrHashLoadBalancer struct {
		eventLoops []Proxy
		size       int
	}
)

// ==================================== Implementation of random load-balancer ====================================

func (lb *randomLoadBalance) add(items []Proxy) error {
	if len(items) == 0 {
		return errors.ErrLoadBalanceParma
	}
	lb.proxySet = append(lb.proxySet, items...)
	return nil
}

func (lb *randomLoadBalance) next(_ string) (Proxy, error) {
	if len(lb.proxySet) == 0 {
		return nil, errors.ErrBalanceProxyIsNil
	}
	lb.nextLoopIndex = rand.Intn(len(lb.proxySet))
	return lb.proxySet[lb.nextLoopIndex], nil
}

// ==================================== Implementation of Round-Robin load-balancer ====================================

func (lb *roundLoadBalancer) add(items []Proxy) error {
	if len(items) == 0 {
		return errors.ErrLoadBalanceParma
	}
	lb.proxySet = append(lb.proxySet, items...)
	lb.size += len(items)
	return nil
}

func (lb *roundLoadBalancer) next(_ string) (Proxy, error) {
	if len(lb.proxySet) == 0 {
		return nil, errors.ErrBalanceProxyIsNil
	}
	lens := len(lb.proxySet)
	if lb.nextLoopIndex >= lens {
		lb.nextLoopIndex = 0
	}
	proxy := lb.proxySet[lb.nextLoopIndex]
	lb.nextLoopIndex += 1
	return proxy, nil
}

// ================================= Implementation of Least-Connections load-balancer =================================

// 加权负载均衡，主要点就是权重： 算法核心概念在于权重之和： 假设ABC三个节点，权重为2，4，8
// 那么总的流量计算方法： 2/14 流量到达A节点
// 那么总的流量计算方法： 4/14 流量到达B节点
// 那么总的流量计算方法： 8/14 流量到达C节点
// 主要影响的计算，每14次流量应该是一个循环才对，也就以为这需要保存每个proxy的本次循环的权重
//

func (lb *weightServerLoadBalancer) add(items []Proxy) error {
	if len(items) == 0 {
		return errors.ErrLoadBalanceParma
	}
	for _, v := range items {
		node := &weightNode{
			proxy:        v,
			sourceWeight: v.Weight(),
			curWeight:    v.Weight(),
		}
		lb.proxySet = append(lb.proxySet, node)
		lb.size ++
	}
	return nil
}

// next returns the eligible event-loop by taking the root node from minimum heap based on Least-Connections algorithm.
func (lb *weightServerLoadBalancer) next(_ string) (Proxy, error) {
	if len(lb.proxySet) == 0 {
		return nil, errors.ErrBalanceProxyIsNil
	}
	if lb.proxySet[lb.curIdx].curWeight > 0 {
		lb.proxySet[lb.curIdx].curWeight -= 1
		return lb.proxySet[lb.curIdx].proxy, nil
	}
	if lb.curIdx >= len(lb.proxySet) {
		lb.curIdx = 0
		lb.flush()
	} else {
		lb.curIdx += 1
	}
	return lb.proxySet[lb.curIdx].proxy, nil
}

func (lb *weightServerLoadBalancer) flush() {
	for _, v := range lb.proxySet {
		tev := v
		tev.curWeight = tev.sourceWeight
	}
}

// ======================================= Implementation of Hash load-balancer ========================================

func (lb *urlAddrHashLoadBalancer) add(items []Proxy) error {
	if len(items) == 0 {
		return errors.ErrLoadBalanceParma
	}
	lb.eventLoops = append(lb.eventLoops, items...)
	lb.size += len(items)
	return nil
}

// hash converts a string to a unique hash code.
func (lb *urlAddrHashLoadBalancer) hash(target string) int {
	v := int(crc32.ChecksumIEEE([]byte(target)))
	if v >= 0 {
		return v
	}
	return -v
}

// next returns the eligible event-loop by taking the remainder of a hash code as the index of event-loop list.
func (lb *urlAddrHashLoadBalancer) next(url string)  (Proxy, error) {
	hashCode := lb.hash(url)
	return lb.eventLoops[hashCode%lb.size],nil
}

