package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ConnectedCallback ...
type ConnectedCallback func(c *Client)

// MessageCallback ...
type MessageCallback func(c *Client, msg mqtt.Message)

// ClientOptions ...
type ClientOptions struct {
	Broker      string
	ClientID    string
	Username    string
	Password    string
	OnConnected ConnectedCallback
	OnMessage   MessageCallback
}

// Client ...
type Client struct {
	opts   *ClientOptions
	client mqtt.Client
}

func (c *Client) onConnected(mc mqtt.Client) {
	// client connected
	if c.opts.OnConnected != nil {
		c.opts.OnConnected(c)
	}
}

func (c *Client) onConnectionLost(mc mqtt.Client, reason error) {
	// client connection lost
}

func (c *Client) onMessageArriaved(mc mqtt.Client, msg mqtt.Message) {
	if c.opts.OnMessage != nil {
		c.opts.OnMessage(c, msg)
	}
}

// MQTT get mqtt client
func (c *Client) MQTT() mqtt.Client {
	return c.client
}

// Close close client
func (c *Client) Close() error {
	c.client.Disconnect(250)
	return nil
}

// NewClient create client instance
func NewClient(cOpts *ClientOptions) (*Client, error) {
	c := &Client{opts: cOpts}
	mOpts := mqtt.NewClientOptions().
		AddBroker(cOpts.Broker).
		SetClientID(cOpts.ClientID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetOnConnectHandler(c.onConnected).
		SetConnectionLostHandler(c.onConnectionLost).
		SetDefaultPublishHandler(c.onMessageArriaved)

	mOpts.SetUsername(cOpts.Username)
	mOpts.SetPassword(cOpts.Password)
	mc := mqtt.NewClient(mOpts)

	start := time.Now().UnixNano()
	statisticer.AddClient(ActionBegin, 0)
	token := mc.Connect()
	if token.Wait() && token.Error() != nil {
		errorf("[%v] failed: %v", cOpts.ClientID, token.Error())
		statisticer.AddClient(ActionFailed, 0)
		return nil, fmt.Errorf("connected failed: %v", token.Error())
	}
	statisticer.AddClient(ActionDone, time.Now().UnixNano()-start)
	c.client = mc
	return c, nil
}
