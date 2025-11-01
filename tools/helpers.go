package tools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
)

// getHostsFromStorage takes a list of names and finds the hosts for those names
func getHostsFromStorage(storageEngine *storage.Engine, names []string) ([]ssh.ClientInfo, error) {
	hosts := make([]ssh.ClientInfo, 0, len(names))
	var notFound []string
	for _, name := range names {
		host, ok := storageEngine.Get(name)
		if !ok {
			notFound = append(notFound, name)
			continue
		}
		hosts = append(hosts, host)
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("no matching hosts for: %s", strings.Join(notFound, ", "))
	}
	return hosts, nil
}

// taskResult is a single result on that host
type taskResult struct {
	Host   string `json:"host"`
	Result string `json:"result"`
	Err    error  `json:"error"`
}

// performTasksOnHosts performs the task on all hosts in parallel
func performTasksOnHosts(hosts []ssh.ClientInfo, task func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error)) map[string]taskResult {
	var wg sync.WaitGroup
	wg.Add(len(hosts))

	var resultsMx sync.Mutex
	results := make(map[string]taskResult, len(hosts))

	for _, host := range hosts {
		go func(host ssh.ClientInfo) {
			defer wg.Done()
			sshClient := ssh.NewClient(&host)
			err := sshClient.Connect()
			if err != nil {
				resultsMx.Lock()
				results[host.Name] = taskResult{Host: host.Name, Err: err}
				resultsMx.Unlock()
				return
			}
			defer sshClient.Close()

			result, err := task(host, sshClient)
			resultsMx.Lock()
			results[host.Name] = taskResult{Host: host.Name, Result: result, Err: err}
			resultsMx.Unlock()
		}(host)
	}
	wg.Wait()

	return results
}
