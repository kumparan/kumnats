package kumnats

import (
	"errors"
	"sync"
	"time"

	"github.com/kumparan/go-lib/utils"

	"github.com/sirupsen/logrus"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/jasonlvhit/gocron"
	"github.com/kumparan/tapao"
	stan "github.com/nats-io/go-nats-streaming"
)

type (
	// NATS :nodoc:
	NATS interface {
		Publish(subject string, value []byte) error
		Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error)
		QueueSubscribe(subject, queueGroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error)
		Close() error
	}

	// Logger :nodoc:
	Logger interface {
		Error(args ...interface{})
		Errorf(format string, args ...interface{})
	}

	// EventType :nodoc:
	EventType string

	// NatsCallback :nodoc:
	NatsCallback func(conn NATS)

	// NATS :nodoc:
	natsImpl struct {
		conn  stan.Conn
		mutex *sync.RWMutex

		stopCh      chan struct{}
		reconnectCh chan struct{}
		wg          *sync.WaitGroup

		info natsInfo

		workerStatus bool
		workerLock   *sync.Mutex

		opts Options
	}

	// natsInfo contains informations that will be use to reconnecting to nats streaming
	natsInfo struct {
		url         string
		clusterID   string
		clientID    string
		stanOptions []stan.Option
		callback    NatsCallback
	}
)

var defaultOptions = Options{
	redisConn:                             nil,
	failedMessagesRedisKey:                "nats:failed-messages",
	deadMessagesRedisKey:                  "nats:dead-messages",
	reconnectInterval:                     500 * time.Millisecond,
	failedMessagePublishIntervalInSeconds: 120,
	logger:                                logrus.New(),
}

// NewNATSWithCallback IMPORTANT! Not to send any stan.NatsURL or stan.SetConnectionLostHandler as options
func NewNATSWithCallback(clusterID, clientID, url string, fn NatsCallback, stanOptions []stan.Option, options ...Option) (NATS, error) {
	nc := &natsImpl{
		reconnectCh:  make(chan struct{}, 1),
		stopCh:       make(chan struct{}),
		wg:           new(sync.WaitGroup),
		workerStatus: false,
		workerLock:   new(sync.Mutex),
		opts:         defaultOptions,
		mutex:        new(sync.RWMutex),
	}

	for _, opts := range options {
		if err := opts(&nc.opts); err != nil {
			return nil, err
		}
	}

	stanOptions = append(stanOptions, stan.SetConnectionLostHandler(func(conn stan.Conn, reason error) {
		select {
		case <-nc.stopCh:
			return
		case nc.reconnectCh <- struct{}{}:

		}
	}))

	nc.info = natsInfo{
		url:         url,
		clusterID:   clusterID,
		clientID:    clientID,
		callback:    fn,
		stanOptions: stanOptions,
	}

	conn, err := connect(clusterID, clientID, url, stanOptions...)
	if err != nil {
		return nil, err
	}

	nc.setConn(conn)
	// Run callback function
	nc.runCallback()
	nc.run()

	return nc, nil
}

// NewNATSMessageHandler a wrapper to standardize how we handle NATS messages
// Payload (arg 0) should always be empty when the method is called. The payload data will later
// parse data from msg.Data.
func NewNATSMessageHandler(payload MessagePayload, retryAttempts int, retryInterval time.Duration, lambda func(payload MessagePayload) error) stan.MsgHandler {
	return func(msg *stan.Msg) {
		logger := logrus.WithField("msg", utils.Dump(msg))
		defer func(logger *logrus.Entry) {
			err := msg.Ack()
			if err != nil {
				logger.Error(err)
			}
		}(logger)

		if msg.Data == nil {
			logger.Error(ErrNilMessagePayload)
			return
		}

		err := payload.ParseFromBytes(msg.Data)
		if err != nil {
			logger.WithField("error-detail", err).Error(ErrBadUnmarshalResult)
			return
		}
		defer logger.WithField("payload", utils.Dump(payload)).Warn("message payload")

		// process payload here
		err = utils.Retry(retryAttempts, retryInterval, func() error {
			return lambda(payload)
		})
		if err != nil {
			logger.WithField("payload", utils.Dump(payload)).Error(ErrGiveUpProcessingMessagePayload)
		}
	}
}

// connect to nats streaming
func connect(clusterID, clientID, url string, options ...stan.Option) (stan.Conn, error) {
	options = append(options, stan.NatsURL(url))
	nc, err := stan.Connect(clusterID, clientID, options...)
	if err != nil {
		return nil, err
	}
	return nc, nil
}

// Close NatsConnection :nodoc:
func (n *natsImpl) Close() error {
	close(n.stopCh)
	n.wg.Wait()

	if n.checkConnIsValid() {
		err := n.conn.Close()
		if err != nil {
			return err
		}
	}

	if n.opts.redisConn != nil {
		err := n.opts.redisConn.Close()
		if err != nil {
			return err
		}
	}
	close(n.reconnectCh)

	return nil
}

// Publish :nodoc:
func (n *natsImpl) Publish(subject string, v []byte) (err error) {
	if n.checkConnIsValid() {
		err = n.conn.Publish(subject, v)
		if err == nil {
			return nil
		}
	}

	if n.opts.redisConn == nil {
		if err != nil {
			return err
		}
		return errors.New("failed publish to nats streaming")
	}

	// Push to redis if failed
	client := n.opts.redisConn.Get()
	defer func() {
		_ = client.Close()
	}()
	b, err := tapao.Marshal(&natsMessageWithSubject{
		Subject: subject,
		Message: v,
	})
	if err != nil {
		return err
	}

	_, err = redigo.Int(client.Do("RPUSH", n.opts.failedMessagesRedisKey, b))
	if err != nil {
		n.opts.logger.Error("failed to RPUSH to redis. redis connection problem")
		return err
	}

	return nil
}

// QueueSubscribe :nodoc:
func (n *natsImpl) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return n.conn.QueueSubscribe(subject, qgroup, cb, opts...)
}

// Subscribe :nodoc:
func (n *natsImpl) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return n.conn.Subscribe(subject, cb, opts...)
}

func (n *natsImpl) setConn(conn stan.Conn) {
	n.mutex.Lock()
	n.conn = conn
	n.mutex.Unlock()
}

func (n *natsImpl) checkConnIsValid() (b bool) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	if n.conn.NatsConn() != nil && n.conn.NatsConn().IsConnected() {
		return true
	}
	return false
}

func (n *natsImpl) run() {
	if n.opts.redisConn != nil {
		s := gocron.NewScheduler()
		s.Every(n.opts.failedMessagePublishIntervalInSeconds).Seconds().Do(n.publishFailedMessageFromRedis)

		n.wg.Add(1)
		go func() {
			defer n.wg.Done()
			c := s.Start()

			<-n.stopCh
			close(c)
		}()
	}

	n.wg.Add(1)
	go n.reconnectWorker()
}

func (n *natsImpl) runCallback() {
	if n.info.callback != nil {
		n.info.callback(n)
	}
}

func (n *natsImpl) checkWorkerStatus() bool {
	n.workerLock.Lock()
	defer n.workerLock.Unlock()
	return n.workerStatus
}

func (n *natsImpl) setWorkerStatus(b bool) {
	n.workerLock.Lock()
	n.workerStatus = b
	n.workerLock.Unlock()
}

func (n *natsImpl) publishFailedMessageFromRedis() {
	isRun := n.checkWorkerStatus()
	if isRun {
		n.opts.logger.Error("worker is already running")
		return
	}
	n.setWorkerStatus(true)
	defer n.setWorkerStatus(false)

	client := n.opts.redisConn.Get()
	defer func() {
		_ = client.Close()
	}()

	for {
		b, err := redigo.Bytes(client.Do("LPOP", n.opts.failedMessagesRedisKey))
		if err != nil && err != redigo.ErrNil {
			n.opts.logger.Error("failed to LPOP from redis. redis connection problem")
			return
		}

		if len(b) == 0 {
			return
		}

		msg := new(natsMessageWithSubject)
		err = tapao.Unmarshal(b, &msg, tapao.FallbackWith(tapao.JSON))
		if err == nil {
			if n.checkConnIsValid() {
				err = n.conn.Publish(msg.Subject, msg.Message)
				if err == nil {
					continue
				}
			}
		}

		if err != nil {
			n.opts.logger.Error("Error : ", err)
		}

		_, err = client.Do("RPUSH", n.opts.deadMessagesRedisKey, b)
		if err != nil {
			n.opts.logger.Error("failed to RPUSH to dead messages key. redis connection problem")
			return
		}
		if err == stan.ErrConnectionClosed {
			return
		}
	}
}

func (n *natsImpl) reconnectWorker() {
	defer n.wg.Done()
	for {
		select {
		case <-n.stopCh:
			return
		case <-n.reconnectCh:
			if n.checkConnIsValid() {
				continue
			}
			conn, err := connect(n.info.clusterID, n.info.clientID, n.info.url, n.info.stanOptions...)
			if err == nil {
				n.setConn(conn)
				n.runCallback()
				continue
			}
			n.opts.logger.Error("failed to reconnect")
			time.Sleep(n.opts.reconnectInterval)

			select {
			case n.reconnectCh <- struct{}{}:
			case <-n.stopCh:
				return
			}
		}
	}
}
