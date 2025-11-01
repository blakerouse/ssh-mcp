package utils

import (
	"fmt"
	"strings"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
)

// HostIdentifier represents a group and name pair
type HostIdentifier struct {
	Group string
	Name  string
}

// ParseHostIdentifiers parses host identifiers in the format "group:name"
func ParseHostIdentifiers(hostStrings []string) ([]HostIdentifier, error) {
	identifiers := make([]HostIdentifier, 0, len(hostStrings))
	for _, hostStr := range hostStrings {
		parts := strings.SplitN(hostStr, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid host identifier format '%s', expected 'group:name'", hostStr)
		}
		identifiers = append(identifiers, HostIdentifier{
			Group: parts[0],
			Name:  parts[1],
		})
	}
	return identifiers, nil
}

// GetHostsFromStorage takes a list of host identifiers and finds the hosts for those identifiers
func GetHostsFromStorage(storageEngine *storage.Engine, identifiers []HostIdentifier) ([]ssh.ClientInfo, error) {
	hosts := make([]ssh.ClientInfo, 0, len(identifiers))
	var notFound []string
	for _, id := range identifiers {
		host, ok := storageEngine.Get(id.Group, id.Name)
		if !ok {
			notFound = append(notFound, fmt.Sprintf("%s:%s", id.Group, id.Name))
			continue
		}
		hosts = append(hosts, host)
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("no matching hosts for: %s", strings.Join(notFound, ", "))
	}
	return hosts, nil
}

// GetHostsFromGroup gets all hosts from a specific group
func GetHostsFromGroup(storageEngine *storage.Engine, group string) ([]ssh.ClientInfo, error) {
	hosts, err := storageEngine.ListGroup(group)
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts from group %s: %w", group, err)
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("no hosts found in group: %s", group)
	}
	return hosts, nil
}
