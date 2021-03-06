/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2018
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

// Publish, Subscribe and Emit package for gopi.Publisher interface
package event

import (
	gopi "github.com/djthorpe/gopi"
)

type PubSub struct {
	subscribers []chan gopi.Event
}

func NewPubSub(capacity int) *PubSub {
	this := new(PubSub)
	this.subscribers = make([]chan gopi.Event, 0, capacity)
	return this
}

func (this *PubSub) Close() {
	for _, subscriber := range this.subscribers {
		if subscriber != nil {
			close(subscriber)
		}
	}
	this.subscribers = nil
}

func (this *PubSub) Subscribe() <-chan gopi.Event {
	if this.subscribers == nil {
		return nil
	}
	subscriber := make(chan gopi.Event)
	this.subscribers = append(this.subscribers, subscriber)
	return subscriber
}

func (this *PubSub) Unsubscribe(subscriber <-chan gopi.Event) {
	for i := range this.subscribers {
		if this.subscribers[i] == subscriber {
			close(this.subscribers[i])
			this.subscribers[i] = nil
		}
	}
}

func (this *PubSub) Emit(evt gopi.Event) {
	for _, subscriber := range this.subscribers {
		if subscriber != nil {
			subscriber <- evt
		}
	}
}
