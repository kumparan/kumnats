package kumnats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/kumparan/tapao"
	natsServer "github.com/nats-io/gnatsd/server"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/nats-streaming-server/server"
)

const (
	clusterName = "my_test_cluster"
	defaultURL  = "nats://localhost:4222"
	clientName  = "test-client"
)

func runServer(clusterName string, port int) *server.StanServer {
	opts := server.GetDefaultOptions()
	opts.ID = clusterName
	s, err := server.RunServerWithOpts(opts, &natsServer.Options{
		Host: "localhost",
		Port: port,
	})
	if err != nil {
		panic(err)
	}
	return s
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func runAnotherServer(port int) *server.StanServer {
	return runServer(clusterName, port)
}

func newRedisConn(url string) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     100,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", url)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func TestMain(t *testing.M) {
	server := runServer(clusterName, 4222)
	defer server.Shutdown()
	m := t.Run()
	os.Exit(m)
}

func TestSafePublish(t *testing.T) {
	n, err := NewNATSWithCallback(clusterName, clientName, defaultURL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer n.Close()

	err = n.SafePublish("test-channel", []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublish(t *testing.T) {
	n, err := NewNATSWithCallback(clusterName, clientName, defaultURL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer n.Close()

	err = n.Publish("test-channel", []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscribe(t *testing.T) {
	n, err := NewNATSWithCallback(clusterName, clientName, defaultURL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer n.Close()

	subject := "test-channel"

	countMsg := 10
	recieveCh := make(chan []byte, countMsg)
	sub, err := n.Subscribe(subject, func(msg *stan.Msg) {
		recieveCh <- msg.Data
	})
	if err != nil {
		t.Fatal(err)
	}

	type msg struct {
		Data int64 `json:"data"`
	}

	for i := 0; i < countMsg; i++ {
		ms := &msg{
			Data: int64(1554775372665126857),
		}
		msgBytes, err := tapao.Marshal(ms)
		if err != nil {
			t.Fatal(err)
		}
		err = n.Publish(subject, msgBytes)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < countMsg; i++ {
		b := <-recieveCh
		msg := new(msg)
		err = tapao.Unmarshal(b, msg)
		if err != nil {
			t.Fatal(err)
		}
		if msg.Data != int64(1554775372665126857) {
			t.Fatal("error")
		}

	}
	sub.Unsubscribe()
}

func TestPublishFailedAndSaveToRedis(t *testing.T) {
	port := 21111
	server := runAnotherServer(port)
	defer server.Shutdown()

	m, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	r := newRedisConn(m.Addr())
	conn, err := NewNATSWithCallback(
		clusterName,
		clientName,
		fmt.Sprintf("localhost:%d", port),
		nil,
		nil,
		WithReconnectInterval(10*time.Millisecond),
		WithRedis(r),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	server.Shutdown()

	type msg struct {
		Data string `json:"data"`
	}

	ms := &msg{
		Data: "test",
	}
	msgBytes, err := json.Marshal(ms)
	if err != nil {
		t.Fatal(err)
	}
	err = conn.SafePublish("test", msgBytes)
	if err != nil {
		t.Fatal(err)
	}

	rClient := r.Get()
	defer rClient.Close()

	b, err := redigo.Bytes(rClient.Do("LPOP", defaultOptions.failedMessagesRedisKey))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), "test") {
		t.Fatal("string bytes should contain message value")
	}
}

func TestReconnectAfterLostConnection(t *testing.T) {
	port := 22222
	server := runAnotherServer(port)
	defer server.Shutdown()

	conn, err := NewNATSWithCallback(clusterName, clientName, fmt.Sprintf("localhost:%d", port), nil, nil, WithReconnectInterval(10*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// shutdown the server
	server.Shutdown()

	v, ok := conn.(*natsImpl)
	if !ok {
		t.Fatal("failed on type assertion")
	}

	if v.checkConnIsValid() {
		t.Fatal("should be not valid")
	}

	server = runAnotherServer(port)
	defer server.Shutdown()

	time.Sleep(1800 * time.Millisecond) // wait for making connection

	if !v.checkConnIsValid() {
		t.Fatal("should be valid")
	}
}

func TestSubscribeAfterLostConnection(t *testing.T) {
	port := 23333
	server := runAnotherServer(port)
	defer server.Shutdown()

	subject := "test"
	recieveCh := make(chan []byte, 2)

	conn, err := NewNATSWithCallback(clusterName, clientName, fmt.Sprintf("localhost:%d", port), func(n NATS) {
		_, err := n.Subscribe(subject, func(msg *stan.Msg) {
			select {
			case recieveCh <- msg.Data:

			}
		})
		if err != nil {
			t.Fatal(err)
		}
	}, nil, WithReconnectInterval(10*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	err = conn.SafePublish(subject, []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	<-recieveCh

	server.Shutdown()

	v, ok := conn.(*natsImpl)
	if !ok {
		t.Fatal("failed on type assertion")
	}

	if v.checkConnIsValid() {
		t.Fatal("should be not valid")
	}

	server = runAnotherServer(port)
	defer server.Shutdown()

	time.Sleep(5 * time.Second) // wait for making connection

	if !v.checkConnIsValid() {
		t.Fatal("should be valid")
	}
	err = conn.SafePublish(subject, []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	<-recieveCh
	close(recieveCh)
}

func TestRunningWorkerAfterLostConnection(t *testing.T) {
	port := 24444
	server := runAnotherServer(port)
	defer server.Shutdown()

	m, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	r := newRedisConn(m.Addr())
	conn, err := NewNATSWithCallback(
		clusterName,
		clientName,
		fmt.Sprintf("localhost:%d", port),
		nil,
		nil,
		WithReconnectInterval(10*time.Millisecond),
		WithRedis(r),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	server.Shutdown()

	type msg struct {
		Data string `json:"data"`
	}

	ms := &msg{
		Data: "test",
	}
	msgBytes, err := json.Marshal(ms)
	if err != nil {
		t.Fatal(err)
	}
	err = conn.SafePublish("test", msgBytes)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := conn.(*natsImpl)
	if !ok {
		t.Fatal("failed on type assertion")
	}

	if v.checkConnIsValid() {
		t.Fatal("should be not valid")
	}

	server = runAnotherServer(port)
	defer server.Shutdown()

	time.Sleep(6 * time.Second) // wait for making connection

	if !v.checkConnIsValid() {
		t.Fatal("should be valid")
	}

	ch := make(chan struct{}, 1)
	output := captureOutput(func() {
		v.publishFailedMessageFromRedis()
		ch <- struct{}{}
	})
	<-ch
	if strings.Contains(output, "abort due to connection problem") {
		t.Fatal("worker should work properly")
	}
}
