package commands

import (
	"sync"

	"github.com/blakerouse/ssh-mcp/ssh"
)

// CommandResult is a single result on that host
type CommandResult struct {
	Host   string `json:"host"`
	Result string `json:"result"`
	Err    error  `json:"error"`
}

// PerformOnHosts performs the command on all hosts in parallel
func PerformOnHosts(hosts []ssh.ClientInfo, command func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error)) map[string]CommandResult {
	var wg sync.WaitGroup
	wg.Add(len(hosts))

	var resultsMx sync.Mutex
	results := make(map[string]CommandResult, len(hosts))

	for _, host := range hosts {
		go func(host ssh.ClientInfo) {
			defer wg.Done()
			sshClient := ssh.NewClient(&host)
			err := sshClient.Connect()
			if err != nil {
				resultsMx.Lock()
				results[host.Name] = CommandResult{Host: host.Name, Err: err}
				resultsMx.Unlock()
				return
			}
			defer sshClient.Close()

			result, err := command(host, sshClient)
			resultsMx.Lock()
			results[host.Name] = CommandResult{Host: host.Name, Result: result, Err: err}
			resultsMx.Unlock()
		}(host)
	}
	wg.Wait()

	return results
}
