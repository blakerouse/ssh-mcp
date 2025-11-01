package tools

// Registry holds all of the defined tools.
var Registry = newRegistry()

type registry struct {
	tools []Tool
}

func newRegistry() *registry {
	return &registry{}
}

// Register registers a new tool.
func (r *registry) Register(tool Tool) {
	r.tools = append(r.tools, tool)
}

// Tools returns all registered tools.
func (r *registry) Tools() []Tool {
	return r.tools
}
