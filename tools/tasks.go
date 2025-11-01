package tools

import (
	"sync"

	"github.com/blakerouse/ssh-mcp/ssh"
)

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
