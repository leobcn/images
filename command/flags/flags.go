// Package flags is low level package for parsing single flag arguments and
// their associated values. It's useful for CLI applications or applications
// that parses os.Args manually.
package flags

import (
	"errors"
	"fmt"
)

// HasFlag checks whether the given flag name is available or not in the
// argument list.
func HasFlag(name string, args []string) bool {
	_, err := ParseValue(name, args)
	if err != nil {
		return false
	}

	return true
}

// IsFlag checks whether the given argument is a valid flag or not
func IsFlag(arg string) bool {
	if _, err := ParseFlag(arg); err != nil {
		return false
	}

	return true
}

// ParseFlag parses a flags name. A flag can be in form of --name=value,
// -name=value, -n=value, or --name, -name=, etc...  If it's a correct flag,
// the name is returned. If not an empty string and an error message is
// returned
func ParseFlag(arg string) (string, error) {
	if arg == "" {
		return "", errors.New("argument is empty")
	}

	if len(arg) == 1 {
		return "", errors.New("argument is too short")
	}

	if arg[0] != '-' {
		return "", errors.New("argument doesn't start with dash")
	}

	numMinuses := 1

	if arg[1] == '-' {
		numMinuses++
		if len(arg) == 2 {
			return "", errors.New("argument is too short")
		}
	}

	name := arg[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return "", fmt.Errorf("bad flag syntax: %s", arg)
	}

	return name, nil
}

// ParseValue parses the value from the given flag. A flag name can be in
// form of name=value, n=value, n=, n.
func parseSingleFlagValue(flag string) (name, value string) {
	for i, r := range flag {
		if r == '=' {
			value = flag[i+1:]
			name = flag[0:i]
		}
	}

	// special case of "n"
	if name == "" {
		name = flag
	}

	return
}

// ParseName parses the given flagName from the args slice and returns the
// value passed to the flag. An example: args: ["--provider", "aws"] will
// return "aws" for the flag name "provider". An empty string and non error
// means the flag is in the boolean form, i.e: ["--provider", "--foo", "bar].
func ParseValue(flagName string, args []string) (string, error) {
	value, _, err := parseFlagAndValue(flagName, args)
	if err != nil {
		return "", err
	}

	return value, nil
}

// FilterFlag filters the given valid flagName with it's associated value (or
// none) from the args. It returns the remaining arguments. If no flagName is
// passed or if the flagName is invalid, remaining arguments are returned
// without any change.
func FilterFlag(flagName string, args []string) []string {
	_, remainingArgs, err := parseFlagAndValue(flagName, args)
	if err != nil {
		return args
	}

	return remainingArgs
}

// parseFlagAndValue is an internal function to parse the flag name, and return
// the value and remaining args.
func parseFlagAndValue(flagName string, args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, errors.New("argument slice is empty")
	}

	if flagName == "" {
		return "", nil, errors.New("flag name is empty")
	}

	for i, arg := range args {
		flag, err := ParseFlag(arg)
		if err != nil {
			continue
		}

		name, value := parseSingleFlagValue(flag)
		if name != flagName && name[0] != flagName[0] {
			continue
		}

		// flag is in the form of "--flagName=value"
		if value != "" {
			// our flag is the first item in the argument list, so just return
			// the remainings
			if i <= 1 {
				return value, args[i+1:], nil
			}

			// flag is between the first and the last, delete it and return the
			// remaining arguments
			return value, append(args[:i], args[i+1:]...), nil
		}

		// no value found yet, check out the next argument. at least two args
		// must be present
		if len(args) < i+1 {
			continue
		}

		// only one flag is passed and it's ours in the form of ["--flagName"]
		if len(args) == 1 {
			return "", args[1:], nil
		}

		// flag is the latest item and has no value, return til the flagName,
		// ["--foo", "bar", "--flagName"]
		if len(args) == i+1 {
			return "", args[:i], nil
		}

		// next argument is a flag i.e: "--flagName --otherFlag", remove our
		// flag and return the remainings
		if IsFlag(args[i+1]) {
			// flag is between the first and the last, delete and return the
			// remaining arguments
			return "", append(args[:i], args[i+1:]...), nil
		}

		// next flag is a value, +2 because the flag is in the form of
		// "--flagName value".  This means we need to remove two items from the
		// slice
		// value := args[i+1]
		return args[i+1], append(args[:i], args[i+2:]...), nil
	}

	return "", nil, fmt.Errorf("argument is not passed to flag: %s", flagName)
}
