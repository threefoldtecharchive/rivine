package datastore

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/types"
)

const (
	subscriberSet      = "s"
	replicationChannel = "replication"

	// The valid commands
	subscribe   = "subscribe"
	unsubscribe = "unsubscribe"

	// 1 second between checks to the replication channel
	pollPeriod = time.Second * 1
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
	// Test the connection
	_, err := client.Ping().Result()
	if err != nil {
		return nil, errors.New("Failed to create redis connection: " + err.Error())
	}

	return &Redis{
		cl: client,
	}, nil
}

// Ping checks the connection to the database by sending a ping
func (rd *Redis) Ping() error {
	_, err := rd.cl.Ping().Result()
	return err
}

// SaveManager saves a managers state in the predefined hset. The field
// is the string representation of the unlockhash
func (rd *Redis) SaveManager(nsm *NamespaceManager) error {
	data, err := nsm.Serialize()
	if err != nil {
		return err
	}
	_, err = rd.cl.HSet(subscriberSet, nsm.Namespace.String(), data).Result()
	return err
}

// GetManagers loads all active managers, stored in the predefined HSet
func (rd *Redis) GetManagers() (map[Namespace]*NamespaceManager, error) {
	result := rd.cl.HGetAll(subscriberSet).Val()
	nsmMap := make(map[Namespace]*NamespaceManager)
	for ns, mgr := range result {
		nsm := &NamespaceManager{}
		if err := nsm.Deserialize([]byte(mgr)); err != nil {
			return nil, err
		}
		if err := nsm.Namespace.LoadString(ns); err != nil {
			return nil, err
		}
		nsmMap[nsm.Namespace] = nsm
	}
	return nsmMap, nil
}

// DeleteManager removes a manager in case it unsubscribes
func (rd *Redis) DeleteManager(nsm *NamespaceManager) error {
	_, err := rd.cl.HDel(subscriberSet, nsm.Namespace.String()).Result()
	return err
}

// StoreData stores data in an HSET
func (rd *Redis) StoreData(ns Namespace, ID DataID, data []byte) error {
	_, err := rd.cl.HSet(ns.String(), string(ID), data).Result()
	return err
}

// DeleteData removes data from an HSET
func (rd *Redis) DeleteData(ns Namespace, ID DataID) error {
	_, err := rd.cl.HDel(ns.String(), string(ID)).Result()
	return err
}

// Subscribe starts a subscription to the
func (rd *Redis) Subscribe(fn SubEventCallback) {
	ps := rd.cl.Subscribe(replicationChannel)
	rd.rch = ps
	go func() {
		ch := ps.Channel()
		for {
			// Try to read from the message channel
			select {
			case msg := <-ch:
				if msg == nil {
					// Read a nil msg, so the channel is closed
					return
				}
				ev, ok := parsePayload(msg)
				if !ok {
					continue
				}
				// Don't waste time here
				go fn(ev)
			default:
				time.Sleep(pollPeriod)
				continue
			}
		}
	}()
}

// Close gracefully closes the database connection
func (rd *Redis) Close() error {
	// Try to close the subscription
	err := rd.rch.Close()
	if err != nil {
		build.Severe("Failed to close channel subscription: ", err)
	}
	// And the db connection
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
