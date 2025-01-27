package main

import "log"

// fire and forget events lmao

/*

type Entry struct {
	key       string
	value     string
	canExpire bool
	TTL       int64
}

Subscribers[conn] -> [..&WriteChanSecondary,..,..,..]
*/

// a single goroutine handles all the events
func eventNotificationHandler() {
	for {
		event := <-EventQueue

		if event.isExpiryEvent {
			SubscribersMutex.Lock()
			subs, exists := Subscribers["~Ex"]
			SubscribersMutex.Unlock()

			if exists {
				for _, ch := range subs {
					*ch <- "[EXPIRY] key: " + event.key + " expired for value: " + event.value
				}
			} else {
				log.Fatal("~Ex not subbed by anyone ever :(")
			}
		} else {
			SubscribersMutex.Lock()
			subs, exists := Subscribers[event.key]
			SubscribersMutex.Unlock()

			if exists {
				for _, ch := range subs {
					*ch <- "[SUBSCRIPTION] key: " + event.key + " has a new value: " + event.value
				}
			}
		}

	}
}
