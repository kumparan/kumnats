package kumnats

import (
	"fmt"
	"os"
	"testing"
	"time"

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

func runAnotherServer(port int) *server.StanServer {
	return runServer(clusterName, port)
}

func TestMain(t *testing.M) {
	server := runServer(clusterName, 4222)
	defer server.Shutdown()
	m := t.Run()
	os.Exit(m)
}

func TestPublish(t *testing.T) {
	n, err := NewNATSWithCallback(clusterName, clientName, defaultURL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer n.Close()

	err = n.Publish("test-channel", "test")
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
	recieveCh := make(chan interface{}, countMsg)
	sub, err := n.Subscribe(subject, func(msg *stan.Msg) {
		recieveCh <- msg.Data
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < countMsg; i++ {
		err = n.Publish(subject, "test")
		if err != nil {
			t.Fatal(err)
		}

	}
	for i := 0; i < countMsg; i++ {
		<-recieveCh
	}
	sub.Unsubscribe()
}

func TestQueueSubscribe(t *testing.T) {
	t.Log("not yet")
}

func TestPublishFailedAndSaveToRedis(t *testing.T) {
	t.Log("not yet")
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

	err = conn.Publish(subject, "test")
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
	err = conn.Publish(subject, "test")
	if err != nil {
		t.Fatal(err)
	}
	<-recieveCh
	close(recieveCh)
}

func TestRunningWorkerAfterLostConnection(t *testing.T) {
	t.Log("not yet")
}
