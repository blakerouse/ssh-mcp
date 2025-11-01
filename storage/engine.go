package storage

import (
	"encoding/json"
	"fmt"

	"github.com/blakerouse/ssh-mcp/ssh"
	badger "github.com/dgraph-io/badger/v4"
)

const hostsPrefix = "host:"

// Engine is the storage engine for SSH connections.
type Engine struct {
	db *badger.DB

	// path to store the database
	path string
}

// NewEngine creates a new storage Engine instance.
func NewEngine(path string) (*Engine, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Disable logging
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	e := &Engine{
		db:   db,
		path: path,
	}
	return e, nil
}

// Close closes the database connection.
func (e *Engine) Close() error {
	if e.db != nil {
		return e.db.Close()
	}
	return nil
}

// Get retrieves the SSH client information for a host.
func (e *Engine) Get(host string) (ssh.ClientInfo, bool) {
	var info ssh.ClientInfo
	key := []byte(hostsPrefix + host)

	err := e.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &info)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ssh.ClientInfo{}, false
		}
		return ssh.ClientInfo{}, false
	}
	return info, true
}

// Set saves the SSH client information for a host.
func (e *Engine) Set(info ssh.ClientInfo) error {
	key := []byte(hostsPrefix + info.Name)
	value, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal client info: %w", err)
	}

	err = e.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
	if err != nil {
		return fmt.Errorf("failed to store client info: %w", err)
	}
	return nil
}

// Delete removes the SSH client information for a host.
func (e *Engine) Delete(host string) error {
	key := []byte(hostsPrefix + host)
	err := e.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		return fmt.Errorf("failed to delete client info: %w", err)
	}
	return nil
}

// List retrieves the names of all hosts.
func (e *Engine) List() ([]ssh.ClientInfo, error) {
	var hosts []ssh.ClientInfo
	prefix := []byte(hostsPrefix)

	err := e.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var info ssh.ClientInfo
				if err := json.Unmarshal(val, &info); err != nil {
					return err
				}
				hosts = append(hosts, info)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %w", err)
	}
	return hosts, nil
}
