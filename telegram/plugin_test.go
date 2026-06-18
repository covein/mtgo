package telegram

import (
	"context"
	"testing"
	"time"
)

type stopCallsRawPlugin struct {
	client *Client
}

func (p *stopCallsRawPlugin) Name() string {
	return "stop_calls_raw"
}

func (p *stopCallsRawPlugin) Start(context.Context, *Client) error {
	return nil
}

func (p *stopCallsRawPlugin) Stop(context.Context) error {
	_ = p.client.Raw()
	return nil
}

func TestStopPluginsDoesNotHoldClientLock(t *testing.T) {
	client, _ := newClientWithMock(t)
	client.Use(&stopCallsRawPlugin{client: client})

	done := make(chan struct{})
	go func() {
		client.stopPlugins(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stopPlugins deadlocked while plugin accessed the client")
	}
}
