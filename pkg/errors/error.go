package errors

import "errors"

var ErrLoadBalanceParma = errors.New("param len 1 at least")
var ErrBalanceProxyIsNil = errors.New("proxy is nil ")