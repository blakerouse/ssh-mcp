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

// makeKey creates a key for storing host information.
// Format: host:group:name
func makeKey(group, name string) []byte {
	return []byte(hostsPrefix + group + ":" + name)
}

// Get retrieves the SSH client information for a host in a group.
func (e *Engine) Get(group, name string) (ssh.ClientInfo, bool) {
	var info ssh.ClientInfo
	key := makeKey(group, name)

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
// The group is taken from info.Group.
func (e *Engine) Set(info ssh.ClientInfo) error {
	if info.Group == "" {
		return fmt.Errorf("group cannot be empty")
	}
	if info.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	key := makeKey(info.Group, info.Name)
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

// Delete removes the SSH client information for a host in a group.
func (e *Engine) Delete(group, name string) error {
	key := makeKey(group, name)
	err := e.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		return fmt.Errorf("failed to delete client info: %w", err)
	}
	return nil
}

// List retrieves all hosts across all groups.
func (e *Engine) List() ([]ssh.ClientInfo, error) {
	return e.listWithPrefix(hostsPrefix)
}

// ListGroup retrieves all hosts in a specific group.
func (e *Engine) ListGroup(group string) ([]ssh.ClientInfo, error) {
	return e.listWithPrefix(hostsPrefix + group + ":")
}

// ListGroups retrieves all unique group names.
func (e *Engine) ListGroups() ([]string, error) {
	groupSet := make(map[string]struct{})
	prefix := []byte(hostsPrefix)

	err := e.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			// Key format: "host:group:name"
			// Extract group from key
			parts := splitKey(key)
			if parts.Group != "" {
				groupSet[parts.Group] = struct{}{}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	groups := make([]string, 0, len(groupSet))
	for group := range groupSet {
		groups = append(groups, group)
	}
	return groups, nil
}

// listWithPrefix is a helper function to list hosts with a given prefix.
func (e *Engine) listWithPrefix(prefix string) ([]ssh.ClientInfo, error) {
	var hosts []ssh.ClientInfo
	prefixBytes := []byte(prefix)

	err := e.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefixBytes
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

// KeyParts represents the parsed components of a storage key.
type KeyParts struct {
	Prefix string // e.g., "host"
	Group  string // e.g., "production"
	Name   string // e.g., "server1"
}

// splitKey splits a key into its components.
// Format: "host:group:name" -> KeyParts{Prefix: "host", Group: "group", Name: "name"}
func splitKey(key string) KeyParts {
	var parts KeyParts
	colonCount := 0
	start := 0

	for i := 0; i < len(key); i++ {
		if key[i] == ':' {
			switch colonCount {
			case 0:
				parts.Prefix = key[start:i]
			case 1:
				parts.Group = key[start:i]
			}
			colonCount++
			start = i + 1
		}
	}

	// The remainder is the name
	if start < len(key) {
		parts.Name = key[start:]
	}

	return parts
}
