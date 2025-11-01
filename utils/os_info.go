package utils

import (
	"fmt"
	"strings"

	"github.com/blakerouse/ssh-mcp/ssh"
)

// GatherOSInfo detects the operating system and gathers relevant system information
func GatherOSInfo(sshClient *ssh.Client) (osRelease string, uname string, err error) {
	// Try to detect the OS by checking if common commands exist
	// First, try Linux/Unix commands
	osReleaseOutput, err := sshClient.Exec("cat /etc/os-release 2>/dev/null || echo ''")
	if err != nil {
		return "", "", fmt.Errorf("failed to check for Linux OS: %w", err)
	}

	// If we got os-release content, it's Linux
	if strings.TrimSpace(string(osReleaseOutput)) != "" {
		unameOutput, err := sshClient.Exec("uname -a")
		if err != nil {
			return "", "", fmt.Errorf("failed to get uname output: %w", err)
		}
		return string(osReleaseOutput), string(unameOutput), nil
	}

	// Try Windows detection with 'ver' command
	verOutput, err := sshClient.Exec("ver 2>nul || echo ''")
	if err == nil && strings.TrimSpace(string(verOutput)) != "" {
		// It's Windows - gather Windows system information
		return gatherWindowsInfo(sshClient)
	}

	// Try PowerShell-based detection as fallback
	psVersion, err := sshClient.Exec("powershell -Command \"$PSVersionTable.PSVersion.ToString()\" 2>nul || echo ''")
	if err == nil && strings.TrimSpace(string(psVersion)) != "" {
		return gatherWindowsInfo(sshClient)
	}

	// Try systeminfo command (works in cmd.exe on Windows)
	systemInfoOutput, err := sshClient.Exec("systeminfo 2>nul | findstr /B /C:\"OS Name\" /C:\"OS Version\" || echo ''")
	if err == nil && strings.TrimSpace(string(systemInfoOutput)) != "" {
		return gatherWindowsInfo(sshClient)
	}

	// If we couldn't detect the OS, return an error
	return "", "", fmt.Errorf("unable to detect operating system - tried Linux and Windows detection methods")
}

// gatherWindowsInfo gathers system information from a Windows host
func gatherWindowsInfo(sshClient *ssh.Client) (osRelease string, uname string, err error) {
	// Use systeminfo for detailed Windows information
	systemInfo, err := sshClient.Exec("systeminfo")
	if err != nil {
		// Fallback to simpler commands
		verOutput, verErr := sshClient.Exec("ver")
		hostnameOutput, hostErr := sshClient.Exec("hostname")

		if verErr != nil || hostErr != nil {
			return "", "", fmt.Errorf("failed to gather Windows system information: %w", err)
		}

		// Format similar to Linux for consistency
		osRelease = fmt.Sprintf("NAME=\"Microsoft Windows\"\nVERSION=\"%s\"", strings.TrimSpace(string(verOutput)))
		uname = fmt.Sprintf("Windows %s", strings.TrimSpace(string(hostnameOutput)))
		return osRelease, uname, nil
	}

	// Parse systeminfo output to extract key information
	systemInfoStr := string(systemInfo)
	lines := strings.Split(systemInfoStr, "\n")

	var osName, osVersion, hostname, architecture string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "OS Name:") {
			osName = strings.TrimSpace(strings.TrimPrefix(line, "OS Name:"))
		} else if strings.HasPrefix(line, "OS Version:") {
			osVersion = strings.TrimSpace(strings.TrimPrefix(line, "OS Version:"))
		} else if strings.HasPrefix(line, "Host Name:") {
			hostname = strings.TrimSpace(strings.TrimPrefix(line, "Host Name:"))
		} else if strings.HasPrefix(line, "System Type:") {
			architecture = strings.TrimSpace(strings.TrimPrefix(line, "System Type:"))
		}
	}

	// Format in a Linux-like style for consistency
	osRelease = fmt.Sprintf("NAME=\"%s\"\nVERSION=\"%s\"\nARCHITECTURE=\"%s\"", osName, osVersion, architecture)
	uname = fmt.Sprintf("Windows %s %s", hostname, architecture)

	return osRelease, uname, nil
}
