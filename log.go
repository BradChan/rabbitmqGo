/*
@Time : 2020/9/17 3:28 下午
@Author : chen
@File : log
@Software: GoLand
*/
package rabbitmqGo

type LogLevel string

const (
	LevelDebug   LogLevel = "debug"
	LevelInfo             = "info"
	LevelWarning          = "warning"
	LevelError            = "error"
)

// Log is used to help select what level logs the application wants to log or
// ignore. Logs are streamed via the `Server.logChan` field which is an optional
// argument. It is provided when calling the `NewServer()` method.
type Log struct {
	Level   LogLevel
	Message string
}
