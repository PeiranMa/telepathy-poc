// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/apache/pulsar-client-go/pulsar"
)

func main() {

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://pulsar-proxy:6650",
	})

	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       "persistent://public/default/latency",
		SubscriptionName:            "consumer-test",
		SubscriptionInitialPosition: pulsar.SubscriptionPositionEarliest,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer consumer.Close()

	// keep message stats
	msgReceived := int64(0)

	// Print stats of the consume rate
	tick := time.NewTicker(10 * time.Second)

	for {
		select {
		case cm, ok := <-consumer.Chan():
			if !ok {
				return
			}
			msgReceived++

			consumer.Ack(cm.Message)
		case <-tick.C:
			currentMsgReceived := atomic.SwapInt64(&msgReceived, 0)

			msgRate := float64(currentMsgReceived) / float64(10)

			log.Infof(`Stats - Consume rate: %6.1f msg/s `,
				msgRate)

		}
	}
}
