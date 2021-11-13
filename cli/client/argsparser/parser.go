package argsparser

import (
	"fmt"
)

const START_COMMAND = "start"
const STOP_COMMAND = "stop"
const QUERY_COMMAND = "query"
const STREAM_COMMAND = "stream"

func GetParams(args []string) (*Parameters, error) {
	argsLen := len(args)

	if argsLen < 2 {
		return nil, fmt.Errorf("invalid parameters %v", args)
	}

	switch args[0] {
	case START_COMMAND:
		return getStartCommandParams(args[1:])
	case STOP_COMMAND:
		return getJobCommandParams(STOP_COMMAND, args[1:])
	case QUERY_COMMAND:
		return getJobCommandParams(QUERY_COMMAND, args[1:])
	case STREAM_COMMAND:
		return getJobCommandParams(STREAM_COMMAND, args[1:])
	}

	return nil, fmt.Errorf("invalid command %v", args)
}

func getJobCommandParams(command string, args []string) (*Parameters, error) {
	params := Parameters{
		CLICommand: command,
	}

	if len(args) != 2 || args[0] != "-j" {
		return nil, fmt.Errorf("invalid parameters for %v command: %v", params.CLICommand, args)
	}

	params.JobID = args[1]

	return &params, nil
}

func getStartCommandParams(args []string) (*Parameters, error) {
	params := Parameters{
		CLICommand: START_COMMAND,
	}

	if len(args) < 2 || args[0] != "-c" {
		return nil, fmt.Errorf("invalid parameters for %v command: %v", params.CLICommand, args)
	}

	params.CommandName = args[1]

	if len(args) >= 4 && args[2] == "-args" {
		params.Arguments = args[3:]
	}

	return &params, nil
}
