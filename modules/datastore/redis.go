package datastore

import (
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/rivine/rivine/types"
)

const (
	replicationChannel = "replication"

	// The valid commands
	subscribe   = "subscribe"
	unsubscribe = "unsubscribe"
)

// Redis wraps a redis connection
type Redis struct {
	cl  *redis.Client
	rch *redis.PubSub // Reference to the replication channel subscription
}

// NewRedis creates a new redis struct which attempts to connect to the specified DB/instance
// The connection always used tcp
func NewRedis(addr, password string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &Redis{
		cl: client,
	}, nil
}

// Ping checks the connection to the database by sending a ping
func (rd *Redis) Ping() error {
	_, err := rd.cl.Ping().Result()
	return err
}

// StoreData stores data in an HSET
func (rd *Redis) StoreData(key string, field string, data []byte) error {
	_, err := rd.cl.HSet(key, field, data).Result()
	return err
}

// DeleteData removes data from an HSET
func (rd *Redis) DeleteData(key string, field string) error {
	_, err := rd.cl.HDel(key, field).Result()
	return err
}

// LoadFieldsForKey returns all field-value mappings in an HSET defined by key
func (rd *Redis) LoadFieldsForKey(key string) (map[string][]byte, error) {
	result := rd.cl.HGetAll(subscriberSet).Val()
	resultMap := make(map[string][]byte)
	for k, v := range result {
		resultMap[k] = []byte(v)
	}
	return resultMap, nil
}

// Subscribe starts a subscription to the replication channel. Once the subscribtion ends
// after a call to Unsubscribe, the channel is also closed
func (rd *Redis) Subscribe(seChan chan<- *SubEvent) {
	ps := rd.cl.Subscribe(replicationChannel)
	rd.rch = ps
	go func() {
		ch := ps.Channel()
		for msg := range ch {
			if msg == nil {
				// Read a nil msg, so the channel is closed
				// Close the channel to the datastore
				close(seChan)
				return
			}
			ev, ok := parsePayload(msg)
			if !ok {
				continue
			}
			seChan <- &ev
		}
	}()
}

// Unsubscribe stops the subsciption on the replication channel
func (rd *Redis) Unsubscribe() error {
	if rd.rch == nil {
		return nil
	}
	return rd.rch.Close()
}

// Close gracefully closes the database connection
func (rd *Redis) Close() error {
	// Try to close the db connection
	return rd.cl.Close()
}

// parsePayload attempts to parse a subscription message to a subevent
func parsePayload(msg *redis.Message) (SubEvent, bool) {
	ev := SubEvent{}
	parts := strings.Split(msg.Payload, ":")
	if len(parts) < 2 {
		return ev, false
	}
	// Get the command
	switch parts[0] {
	case subscribe:
		ev.Action = SubStart
		break
	case unsubscribe:
		ev.Action = SubEnd
		break
	default:
		return ev, false
	}
	// Set the namepsace
	if err := ev.Namespace.LoadString(string(parts[1])); err != nil {
		// Malformed namespace
		return ev, false
	}
	// If there is more to come, check if we can add a starttime
	// Setting a starttime for sub end is useless, but it doesnt matter
	// that we set it
	if len(parts) > 2 {
		ut, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			return ev, true
		}
		ev.Start = types.Timestamp(ut)
	}
	// Ignore other segments
	return ev, true
}
