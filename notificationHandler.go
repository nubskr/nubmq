package main

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
func evenNotificationHandler() {
	for {
		event := <-EventQueue

		SubscribersMutex.Lock()
		subs, exists := Subscribers[event.key]
		SubscribersMutex.Unlock()

		if exists {
			for _, ch := range subs {
				*ch <- "[SUBSCRIPTION] key: " + event.key + " has a new value: " + event.value
			}
		} else {
		}
	}
}
