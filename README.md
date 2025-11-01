# SSH AI
SSH AI is an MCP server that runs locally on your host that provides the ability to add and remove
hosts that you can SSH into and perform commands.

## Tools

- Add host
  - Adds a host
- Remove host
  - Removes a host
- Get OS Info
  - Shows the OS information of the hosts
- Update OS Info
  - Updates the cached OS information of the hosts
- Perform Command
  - Performs the command on the provided hosts

## Limitations

- Only supports Linux
- Only supports username/password SSH
- Ignores host keys


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

First add a host that you want to be able to interact with:

`add host with name <name> connecting with ssh://<USER>:<PASS>@<IP>`

You can then list the hosts available with:

`list my hosts`

You can add another host with same command again (multiple hosts will perform the same
command on the provided hosts):

`add host with name <name> connecting with ssh://<USER>:<PASS>@<IP>`

Ask for the OS information for all added hosts:

`show OS information for all hosts`

Ask for it to check if any of the hosts need to be updated:

`do any of my hosts need to be updated`

Ask it to upgrade a specific host (or all, it will do it at the same time):

`upgrade host <name>`
