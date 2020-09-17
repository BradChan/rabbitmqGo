/*
@Time : 2020/9/17 3:27 下午
@Author : chen
@File : exchange
@Software: GoLand
*/
package rabbitmqGo

// ExchangeConfig is used by both producer and consumer applications.
type ExchangeConfig struct {
	// Name defines the name of the exchange. Required.
	Name string
	// Type defines the type of the exchange. Required.
	Type string
}
