# SSH MCP
SSH MCP is an MCP server that runs locally on your host that provides the ability to manage and interact with remote SSH hosts. It supports organizing hosts into groups and executing commands across multiple hosts simultaneously.

## Tools

- **add_host** - Adds a new host to the SSH configuration. Username and password are optional in the connection string - if not provided, the current user and SSH agent will be used for authentication.
- **remove_host** - Removes a host from the SSH configuration by group and name.
- **get_groups** - Retrieves the list of all groups from the SSH configuration.
- **get_hosts** - Retrieves the list of hosts from the SSH configuration. Can optionally filter by group.
- **get_os_info** - Retrieves the cached operating system information from hosts. You can specify individual hosts or an entire group.
- **update_os_info** - Updates the cached operating system information. You can specify individual hosts or an entire group.
- **perform_command** - SSH into a remote machine and executes a command. You can specify individual hosts or an entire group.

## Features

- **Group-based organization** - Organize hosts into groups for easier management
- **Multiple authentication methods** - Supports password, SSH agent, and SSH key files (~/.ssh/id_rsa, id_ed25519, etc.)
- **Secure host verification** - Uses ~/.ssh/known_hosts for host key verification with automatic host addition
- **Concurrent execution** - Execute commands across multiple hosts simultaneously
- **Persistent storage** - Uses BadgerDB for efficient local storage

## Limitations

- Remote hosts must be Linux-based (OS detection commands assume Linux)
- SSH agent support is Unix-only (SSH_AUTH_SOCK) - password and key file authentication work on all platforms


## How to Setup

Checkout the repository:

```shell
$ git clone https://github.com/blakerouse/ssh-mcp
```

Build the binary:

```shell
$ go build .
```

Update the MCP configuration for Claude Desktop:

```
{
  "mcpServers": {
    "ssh": {
      "command": "<PATH_TO_BUILT_SSHAI_BINARY>",
      "args": ["--storage", "<PATH_TO_STORE_HOSTS>"]
    }
  }
}
```

Restart Claude Desktop

## How to Use

### Adding Hosts

Add a host to a group (multiple formats supported - `ssh://` prefix is optional):

```
add host to production group connecting with 10.0.1.5
add host named web01 to production group connecting with user@10.0.1.5:2222
add host to staging group connecting with user:pass@10.0.1.10
```

### Listing Groups and Hosts

List all groups:
```
show me all groups
```

List all hosts:
```
list all my hosts
```

List hosts in a specific group:
```
show me hosts in production group
```

### Getting OS Information

Get OS info for all hosts in a group:
```
show OS information for production group
```

Get OS info for specific hosts:
```
show OS information for production:web01 and production:web02
```

### Executing Commands

Run a command on all hosts in a group:
```
run "uptime" on production group
check disk space on staging group
```

Run a command on specific hosts:
```
run "systemctl status nginx" on production:web01 and production:web02
```

### Updating OS Information

Update cached OS information:
```
update OS information for production group
refresh OS info for staging:db01
```

### Managing Hosts

Remove a host:
```
remove production:web01
remove host web02 from staging group
```
