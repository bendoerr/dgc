package dgc

import "strings"

// Command represents a simple command
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Example     string
	Flags       []string
	IgnoreCase  bool
	SubCommands []*Command
	RateLimiter RateLimiter
	Handler     ExecutionHandler
}

// GetSubCmd returns the sub command with the given name if it exists
func (command *Command) GetSubCmd(name string) *Command {
	for _, subCommand := range command.SubCommands {
		if subCommand.Name == name || stringArrayContains(subCommand.Aliases, name, subCommand.IgnoreCase) {
			return subCommand
		}
	}
	return nil
}

// trigger triggers the given command
func (command *Command) trigger(ctx *Ctx) {
	// Check if the first argument matches a sub command
	if len(ctx.Arguments.arguments) > 0 {
		argument := ctx.Arguments.Get(0).Raw()
		subCommand := command.GetSubCmd(argument)
		if subCommand != nil {
			// Define the arguments for the sub command
			arguments := ParseArguments("")
			if ctx.Arguments.Amount() > 1 {
				arguments = ParseArguments(strings.Join(strings.Split(ctx.Arguments.Raw(), " ")[1:], " "))
			}

			// Trigger the sub command
			subCommand.trigger(&Ctx{
				Session:       ctx.Session,
				Event:         ctx.Event,
				Arguments:     arguments,
				CustomObjects: ctx.CustomObjects,
				Router:        ctx.Router,
				Command:       subCommand,
			})
			return
		}
	}

	// Check if the user is being rate limited
	if command.RateLimiter != nil && !command.RateLimiter.NotifyExecution(ctx) {
		return
	}

	// Run all middlewares assigned to this command
	for _, flag := range command.Flags {
		for _, middleware := range ctx.Router.Middlewares[flag] {
			if !middleware(ctx) {
				return
			}
		}
	}

	// Handle this command if the first argument matched no sub command
	command.Handler(ctx)
}
