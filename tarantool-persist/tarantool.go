package persist

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tarantool/go-tarantool"
)

// Tarantool database related errors
var (
	ErrDuplicateKey = errors.New("Key already exists in space")
)

// TarantoolClient wraps tarantool.Connection
type TarantoolClient struct {
	connection *tarantool.Connection
}

// NewTarantoolClient creates a new TarantoolClient and makes connection to db
func NewTarantoolClient() *TarantoolClient {
	server := "127.0.0.1:3301"
	opts := tarantool.Opts{
		Timeout:       8000 * time.Millisecond,
		Reconnect:     5 * time.Second,
		MaxReconnects: 20,
	}
	client, err := tarantool.Connect(server, opts)
	if err != nil {
		log.Fatalf("Failed to connect: %s", err.Error())
	}

	res, err := client.Ping()
	log.Print(res)
	if err != nil {
		log.Printf("Failed to ping client: %s", err.Error())
	}

	return &TarantoolClient{
		connection: client,
	}
}

// Eval wraps tarantool client.connection.Eval function and handles errors
func (client *TarantoolClient) Eval(expr string, params interface{}) ([]interface{}, error) {
	resp, err := client.connection.Eval(expr, params)
	if err != nil {
		return nil, fmt.Errorf("Error accessing database with error %s", err.Error())
	}
	return resp.Data, nil
}

// Call tarantool client.connection.Call function and handles errors
func (client *TarantoolClient) Call(function string, params []interface{}) ([]interface{}, error) {
	resp, err := client.connection.Call(function, params)
	if err != nil {
		return nil, fmt.Errorf("Error accessing database with error %s", err.Error())
	}
	return resp.Data, nil
}

// Get wraps tarantool client.connection.Select function and handles errors
func (client *TarantoolClient) Get(space, index string, offset, limit, iterator uint32, key interface{}) ([]interface{}, error) {
	tSpace := client.connection.Schema.Spaces[space]
	resp, err := client.connection.Select(tSpace, index, offset, limit, iterator, []interface{}{key})
	if err != nil {
		return nil, fmt.Errorf("Error accessing database with error %s", err.Error())
	}
	return resp.Data, nil
}

// Insert wraps tarantool client.connection.Insert function and handles errors
func (client *TarantoolClient) Insert(space string, value []interface{}) ([]interface{}, error) {
	tSpace := client.connection.Schema.Spaces[space]
	resp, err := client.connection.Insert(tSpace, value)
	if err != nil && resp.Code == 3 {
		return nil, ErrDuplicateKey
	}
	if err != nil {
		return nil, fmt.Errorf("Error accessing database with error %s", err.Error())
	}
	return resp.Data, nil
}

// Insert wraps tarantool client.connection.Insert function and handles errors
func (client *TarantoolClient) Upsert(space string, key interface{}, value interface{}) ([]interface{}, error) {
	tSpace := client.connection.Schema.Spaces[space]
	resp, err := client.connection.Upsert(tSpace, []interface{}{key, value}, []interface{}{[]interface{}{"+", 1, 1}})
	if err != nil && resp.Code == 3 {
		return nil, ErrDuplicateKey
	}
	if err != nil {
		return nil, fmt.Errorf("Error accessing database with error %s", err.Error())
	}
	return resp.Data, nil
}

// Update wraps tarantool client.connection.Update function and handles errors
func (client *TarantoolClient) Update(space, index string, key interface{}, value interface{}) ([]interface{}, error) {
	tSpace := client.connection.Schema.Spaces[space]
	resp, err := client.connection.Update(tSpace, index, key, value)
	if err != nil {
		return nil, fmt.Errorf("Error accessing database with error %s", err.Error())
	}
	return resp.Data, nil
}

// Delete wraps tarantool client.connection.Delete function and handles errors
func (client *TarantoolClient) Delete(space, index string, key interface{}) ([]interface{}, error) {
	tSpace := client.connection.Schema.Spaces[space]
	resp, err := client.connection.Delete(tSpace, index, key)
	if err != nil {
		return nil, fmt.Errorf("Error accessing database with error %s", err.Error())
	}
	return resp.Data, nil
}

// CreateSpace wraps tarantool client.connection.eval function to create a space and handles errors
func (client *TarantoolClient) CreateSpace(space string, fieldCount int) error {
	_, err := client.connection.Eval(fmt.Sprintf("box.schema.space.create('%s', { if_not_exists = true , temporary=true, field_count = %d })", space, fieldCount), []interface{}{})
	if err != nil {
		return fmt.Errorf("Error creating UnlockHashSpace, %s", err.Error())
	}
	return nil
}

// CreateIndex wraps tarantool client.connection.eval function to create an index on a space and handles errors
func (client *TarantoolClient) CreateIndex(space string, indexName string, fieldIndex uint16, fieldType string, indexType string) error {
	_, err := client.connection.Eval(fmt.Sprintf("box.space.%s:create_index('%s', { if_not_exists = true, parts = {%v, '%s'}, type='%s'})", space, indexName, fieldIndex, fieldType, indexType), []interface{}{})
	if err != nil {
		return fmt.Errorf("Error creating index id on UnlockHashSpace, %s", err.Error())
	}
	return nil
}

// Close connection
func (client *TarantoolClient) Close() error {
	return client.Close()
}
