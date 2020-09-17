/*
@Time : 2020/9/17 3:29 下午
@Author : chen
@File : server
@Software: GoLand
*/
package rabbitmqGo

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type Server struct {
	mutex     *sync.RWMutex
	conn      *amqp.Connection
	config    ConnectionConfig
	logChan   chan Log
	consumers []Consumer
	channels  map[string]*amqp.Channel
}

// New returns `Server` pointer type with an live AMQP connection attached to
// it.
//
// The optional `logChan` argument helps you get back package level logs. If it
// is to be utilised, you must use an unbuffered `Log` channel and read from it
// right after creating it. Failing to read will prevent the reconnection
// feature from establishing a new connection and possible unexpected issues.
func NewServer(config ConnectionConfig, logChan chan Log) (*Server, error) {
	if config.ReconInterval == 0 {
		return nil, fmt.Errorf("reconnection interval must be above 0")
	}

	srv := &Server{
		mutex:    &sync.RWMutex{},
		config:   config,
		logChan:  logChan,
		channels: make(map[string]*amqp.Channel),
	}
	if err := srv.connect(); err != nil {
		return nil, err
	}

	return srv, nil
}

// Shutdown closes the AMQP connection.
func (s *Server) Shutdown() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	}

	return nil
}

// Setup declares all the necessary components of the broker that is needed for
// producers and consumers.
func (s *Server) Setup(exchanges []ExchangeConfig, queues []QueueConfig) error {
	chn, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	defer chn.Close()

	for _, exchange := range exchanges {
		if err := chn.ExchangeDeclare(
			exchange.Name,
			exchange.Type,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("setup: exchange declare: %w", err)
		}
	}

	for _, queue := range queues {
		args := amqp.Table{"x-queue-mode": queue.Mode}
		if queue.DLX != "" {
			args["x-dead-letter-exchange"] = queue.DLX
		}

		if _, err := chn.QueueDeclare(
			queue.Name,
			true,
			false,
			false,
			false,
			args,
		); err != nil {
			return fmt.Errorf("setup: queue declare: %w", err)
		}

		if err := chn.QueueBind(
			queue.Name,
			queue.Binding,
			queue.Exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("setup: queue bind: %w", err)
		}
	}

	return nil
}

// RegisterConsumers first registers all consumers and then runs their workers.
func (s *Server) RegisterConsumers(consumers []Consumer) error {
	s.consumers = consumers

	if err := s.runConsumerWorkers(s.consumers); err != nil {
		return fmt.Errorf("register consumers: %w", err)
	}

	return nil
}

// PublishOnNewChannel publishes a message on a new channel.
//
// Every time this method is called a new channel is opened and closed right
// after the use. This has a negative impact on the application performance.
//
// The advantage of using a new channel for each publishing is that, it allows
// message delivery confirmation. It is possible for published message to not
// reach the exchange, queue or the server for any reason. The lack of an error
// on the publishing does not necessarily mean that the server has received the
// published message either.
//
// The disadvantage is obviously a new channel is created for each publishing
// and closed right after the use. This will result in considerably slower
// operations and higher usage of system resources such as high channel churn.
// The disadvantage becomes reality if it was used by fairly busy producers.
//
// If the message delivery confirmation is a "must have" feature for your use
// case you have no other choice but use this method. Otherwise always prefer
// the `PublishOnReservedChannel()` method.
//
// Ref: https://www.rabbitmq.com/channels.html
// High Channel Churn
func (s *Server) PublishOnNewChannel(publishing amqp.Publishing, config ProducerConfig) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	chn, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("publish on new channel: get channel: %w", err)
	}
	defer chn.Close()

	if err := chn.Confirm(false); err != nil {
		return fmt.Errorf("publish on new channel: confirm mode: %w", err)
	}

	err = chn.Publish(config.ExchangeName, config.RoutingKey, true, false, publishing)
	if err != nil {
		return fmt.Errorf("publish on new channel: publish: %w", err)
	}

	select {
	case ntf := <-chn.NotifyPublish(make(chan amqp.Confirmation, 1)):
		if !ntf.Ack {
			return errors.New("publish on new channel: failed to confirm publishing")
		}
	case <-chn.NotifyReturn(make(chan amqp.Return)):
		return errors.New("publish on new channel: failed to route publishing")
	}

	return nil
}

// PublishOnReservedChannel publishes a message on previously reserved channel
// on behalf of the producers.
//
// The reserved channels are not closed as they are meant to be long-lived and
// reused for multiple publishing. This has a positive impact on the application
// performance.
//
// The advantage of using a reserved channel is that, each producer uses its own
// reserved channel for each publishing. This will result in considerably
// faster operations and less usage of system resources such as low channel
// churn.
//
// The disadvantage is that, it will not allow message delivery confirmation. If
// you want to what we mean by the message delivery confirmation, please read
// the `PublishOnNewChannel()` method.
//
// If the message delivery confirmation is not important for your use case,
// always prefer this method over `PublishOnNewChannel()` method.
//
// Ref: https://www.rabbitmq.com/channels.html
// High Channel Churn
func (s *Server) PublishOnReservedChannel(publishing amqp.Publishing, config ProducerConfig) error {
	chn, err := s.reservedChannel(config.Name)
	if err != nil {
		return fmt.Errorf("publish on reserved channel: %w", err)
	}

	err = chn.Publish(config.ExchangeName, config.RoutingKey, false, false, publishing)
	if err != nil {
		return fmt.Errorf("publish on reserved channel: publish: %w", err)
	}

	return nil
}

// reservedChannel returns an existing channel for a producer.
//
// If the given producer name does not yet have an channel exist in the reserved
// channel pool, a new channel is created and reserved for later use.
//
// All the reserved channels have a channel listeners `producerChannelListener`
// attached to them so if the channel is closed for any given reason, the
// listener calls this method in order to recreate one.
func (s *Server) reservedChannel(producerName string) (*amqp.Channel, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if chn, ok := s.channels[producerName]; ok {
		return chn, nil
	}

	chn, err := s.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("reserved channel: get channel: %w", err)
	}
	s.channels[producerName] = chn

	go s.producerChannelListener(chn, producerName)

	return chn, nil
}

// runConsumerWorkers runs all the workers that are linked to the given
// consumers.
//
// Each individual consumer gets its own dedicated channel and this channel is
// shared between all its workers. e.g., there are two consumers and each have
// two workers attached to them. We would have a total of two open channels in
// the broker. Given that the workers are run as goroutines and the goroutines
// are not threads, there is no point of using a channel per worker.
//
// All the channels have a channel listeners `consumerChannelListener`
// attached to them so if the channel is closed for any given reason, the
// listener calls `runConsumerWorkers()` method in order to rerun all the
// workers of the consumer.
//
// Ref: https://www.rabbitmq.com/api-guide.html
// Channels and Concurrency Considerations (Thread Safety)
func (s *Server) runConsumerWorkers(consumers []Consumer) error {
	op := "run consumer workers"

	for _, consumer := range consumers {
		consumer := consumer
		config := consumer.Config()

		chn, err := s.conn.Channel()
		if err != nil {
			return fmt.Errorf("%s: get channel: %s: %w", op, config.Name, err)
		}

		if err := chn.Qos(config.PrefetchCount, 0, false); err != nil {
			return fmt.Errorf("%s: qos channel: %s: %w", op, config.Name, err)
		}

		if config.WorkerCount < 1 {
			return fmt.Errorf("%s: insufficient worker count: %s-%d/0", op, config.Name, config.WorkerCount)
		}

		go s.consumerChannelListener(chn, consumer)

		for i := 1; i <= config.WorkerCount; i++ {
			i := i

			go func() {
				messages, err := chn.Consume(
					config.Name,
					fmt.Sprintf("%s (%d/%d)", config.ConsumerName, i, config.WorkerCount),
					false,
					false,
					false,
					false,
					nil,
				)
				if err != nil {
					s.log(Log{
						Level: LevelError,
						Message: fmt.Sprintf(
							"%s: consume channel: %s-%d/%d: %v", op, config.Name, i, config.WorkerCount, err,
						),
					})
				}

				s.log(Log{
					Level:   LevelInfo,
					Message: fmt.Sprintf("%s: run %s-%s-%d/%d", op, config.Name, config.ConsumerName, i, config.WorkerCount),
				})

				consumer.Consume(messages, i)

				s.log(Log{
					Level:   LevelInfo,
					Message: fmt.Sprintf("%s: stop %s-%d/%d", op, config.Name, i, config.WorkerCount),
				})
			}()
		}
	}

	return nil
}

// connect establishes a new connection based on the required schema.
func (s *Server) connect() error {
	if s.config.Schema == "amqp" {
		return s.connectAMQP()
	}

	return s.connectAMQPS()
}

// connectAMQP establishes a new AMQP connection only if there is not one at the
// application bootstrap.
//
// However, this method will also be called by the `connectionListener()` method
// behind the scene as many times as required when the connection goes down.
// Hence the reason why it is also responsible for rerunning the consumer
// workers. Otherwise, workers would not be up and running after reconnection.
func (s *Server) connectAMQP() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var (
		err error
		url string
	)

	if s.config.Username != "" || s.config.Password != "" {
		url = fmt.Sprintf(
			"amqp://%s:%s@%s:%s/%s",
			s.config.Username,
			s.config.Password,
			s.config.Host,
			s.config.Port,
			s.config.VHost,
		)
	} else {
		url = fmt.Sprintf(
			"amqp://%s:%s/%s",
			s.config.Host,
			s.config.Port,
			s.config.VHost,
		)
	}

	s.conn, err = amqp.DialConfig(url,
		amqp.Config{
			Properties: amqp.Table{"connection_name": s.config.Name},
		},
	)
	if err != nil {
		return fmt.Errorf("connect amqp: %w", err)
	}

	if err := s.runConsumerWorkers(s.consumers); err != nil {
		return fmt.Errorf("connect amqp: %w", err)
	}

	go s.connectionListener()

	return nil
}

// connectAMQPS establishes a new AMQPS connection only if there is not one at
// the application bootstrap.
//
// However, this method will also be called by the `connectionListener()` method
// behind the scene as many times as required when the connection goes down.
// Hence the reason why it is also responsible for rerunning the consumer
// workers. Otherwise, workers would not be up and running after reconnection.
func (s *Server) connectAMQPS() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	tlsCnf := &tls.Config{}

	if s.config.CACert != nil {
		tlsCnf.RootCAs = x509.NewCertPool()
		tlsCnf.RootCAs.AppendCertsFromPEM(s.config.CACert)
	}

	if cert, err := tls.X509KeyPair(s.config.ClientCert, s.config.ClientKey); err == nil {
		tlsCnf.Certificates = append(tlsCnf.Certificates, cert)
	}

	var (
		err error
		url string
	)

	if s.config.Username != "" || s.config.Password != "" {
		url = fmt.Sprintf(
			"amqps://%s:%s@%s:%s/%s",
			s.config.Username,
			s.config.Password,
			s.config.Host,
			s.config.Port,
			s.config.VHost,
		)
	} else {
		url = fmt.Sprintf(
			"amqps://%s:%s/%s",
			s.config.Host,
			s.config.Port,
			s.config.VHost,
		)
	}

	s.conn, err = amqp.DialConfig(url,
		amqp.Config{
			Properties:      amqp.Table{"connection_name": s.config.Name},
			TLSClientConfig: tlsCnf,
		},
	)
	if err != nil {
		return fmt.Errorf("connect amqps: %w", err)
	}

	if err := s.runConsumerWorkers(s.consumers); err != nil {
		return fmt.Errorf("connect amqps: %w", err)
	}

	go s.connectionListener()

	return nil
}

// connectionListener listens on the closed connection notifications and
// attempts to reestablish a new connection by calling the `connect()` method.
// However, if the connection was closed explicitly, nothing shall be done.
//
// Total reconnection attempts and intervals are configured within the
// `ConnectionConfig` struct. For the infinite attempts, the `ReconAttempt`
// option must be set to `0`.
func (s *Server) connectionListener() {
	err := <-s.conn.NotifyClose(make(chan *amqp.Error))
	if err != nil {
		op := "connection listener"

		s.log(Log{
			Level:   LevelWarning,
			Message: fmt.Sprintf("%s: closed: %v", op, err),
		})

		ticker := time.NewTicker(s.config.ReconInterval)
		defer ticker.Stop()

		var i int
		for range ticker.C {
			i++

			s.log(Log{
				Level:   LevelDebug,
				Message: fmt.Sprintf("%s: reconnection attempt: %d/%d", op, i, s.config.ReconAttempt),
			})

			if err := s.connect(); err == nil {
				s.log(Log{
					Level:   LevelInfo,
					Message: fmt.Sprintf("%s: reconnected: %d/%d", op, i, s.config.ReconAttempt),
				})
				return
			}

			if i == s.config.ReconAttempt {
				s.log(Log{
					Level:   LevelError,
					Message: fmt.Sprintf("%s: reconnection failed: %d/%d", op, i, s.config.ReconAttempt),
				})
				return
			}
		}
	}

	s.log(Log{
		Level:   LevelInfo,
		Message: fmt.Sprintf("connection listener: explicetly closed the connection"),
	})
}

// producerChannelListener listens on the closed reserved channel notifications
// and removes the channel from the pool.
//
// Once removed, the very first call to the `PublishOnReservedChannel()` method
// will help recreate a new channel.
func (s *Server) producerChannelListener(chn *amqp.Channel, producerName string) {
	err := <-chn.NotifyClose(make(chan *amqp.Error))
	if err != nil {
		s.log(Log{
			Level:   LevelWarning,
			Message: fmt.Sprintf("producer channel listener: closed: %s: %v", producerName, err),
		})
	} else {
		s.log(Log{
			Level:   LevelWarning,
			Message: fmt.Sprintf("producer channel listener: closed: %s", producerName),
		})
	}

	s.mutex.Lock()
	delete(s.channels, producerName)
	s.mutex.Unlock()
}

// consumerChannelListener listens on the closed consumer channel notifications
// and reruns the consumer workers with `runConsumerWorkers()` method. However,
// if the connection was closed explicitly, nothing shall be done.
func (s *Server) consumerChannelListener(chn *amqp.Channel, consumer Consumer) {
	err := <-chn.NotifyClose(make(chan *amqp.Error))
	if err != nil && err.Code == amqp.ConnectionForced {
		return
	}

	if err != nil {
		s.log(Log{
			Level:   LevelWarning,
			Message: fmt.Sprintf("consumer channel listener: closed: %s: %v", consumer.Config().Name, err),
		})
	} else {
		s.log(Log{
			Level:   LevelWarning,
			Message: fmt.Sprintf("consumer channel listener: closed: %s", consumer.Config().Name),
		})
	}

	if err := s.runConsumerWorkers([]Consumer{consumer}); err != nil {
		s.log(Log{
			Level:   LevelError,
			Message: fmt.Sprintf("consumer channel listener: %v", err),
		})
	}
}

// log sends log messages to the log channel if not nil.
func (s *Server) log(log Log) {
	if s.logChan == nil {
		return
	}

	s.logChan <- log
}
