/*
@Time : 2020/9/17 3:27 下午
@Author : chen
@File : connection
@Software: GoLand
*/
package rabbitmqGo

import "time"

type ConnectionConfig struct {
	// Name is used to name the connection. It provides visual cues to which
	// connection belongs to which application in Management Plugin. Optional.
	Name string
	// Schema segment of the AMQP URI string. Required.
	Schema string
	// Username segment of the AMQP URI string. Optional.
	Username string
	// Password segment of the AMQP URI string. Optional.
	Password string
	// Host segment of the AMQP URI string. Required.
	Host string
	// Port segment of the AMQP URI string. Required.
	Port string
	// VHost segment of the AMQP URI string. Optional.
	VHost string
	// ReconAttempt is used to define maximum amount of reconnection attempts.
	// If set to `0` attempts will be infinite. Optional.
	ReconAttempt int
	// ReconInterval defines the equal intervals between each reconnection
	// attempts. Required.
	ReconInterval time.Duration
	// CACert represents Certificate Authority (CA) certificate. Optional.
	CACert []byte
	// ClientCert represents Client certificate. Optional.
	ClientCert []byte
	// ClientCert represents Client key. Optional.
	ClientKey []byte
}
