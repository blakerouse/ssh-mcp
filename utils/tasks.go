package utils

import (
	"sync"

	"github.com/blakerouse/ssh-mcp/ssh"
)

// TaskResult is a single result on that host
type TaskResult struct {
	Host   string `json:"host"`
	Result string `json:"result"`
	Err    error  `json:"error"`
}

// PerformTasksOnHosts performs the task on all hosts in parallel
func PerformTasksOnHosts(hosts []ssh.ClientInfo, task func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error)) map[string]TaskResult {
	var wg sync.WaitGroup
	wg.Add(len(hosts))

	var resultsMx sync.Mutex
	results := make(map[string]TaskResult, len(hosts))

	for _, host := range hosts {
		go func(host ssh.ClientInfo) {
			defer wg.Done()
			sshClient := ssh.NewClient(&host)
			err := sshClient.Connect()
			if err != nil {
				resultsMx.Lock()
				results[host.Name] = TaskResult{Host: host.Name, Err: err}
				resultsMx.Unlock()
				return
			}
			defer sshClient.Close()

			result, err := task(host, sshClient)
			resultsMx.Lock()
			results[host.Name] = TaskResult{Host: host.Name, Result: result, Err: err}
			resultsMx.Unlock()
		}(host)
	}
	wg.Wait()

	return results
}
