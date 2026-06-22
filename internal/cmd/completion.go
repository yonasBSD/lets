package cmd

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/lets-cli/lets/internal/config/config"
	"github.com/lets-cli/lets/internal/docopt"
	"github.com/spf13/cobra"
)

const zshCompletionText = `#compdef lets

LETS_EXECUTABLE=lets

function _lets {
	local state

	_arguments -C -s \
		"(-c --config)"{-c,--config}"[config file (default is lets.yaml)]:config file:_files" \
		"(-E --env)"{-E,--env}"[set env variable KEY=VALUE]:env var:" \
		"--only[run only specified command(s)]:command:" \
		"--exclude[run all but excluded command(s)]:command:" \
		"(-d --debug -dd)"{-d,--debug}"[show debug logs]" \
		"-dd[show very verbose debug logs]" \
		"--all[show all commands]" \
		"--init[create lets.yaml in current folder]" \
		"1: :->cmds" \
		'*::arg:->args'

	case $state in
		cmds)
			_lets_commands
			;;
		args)
			local cmd=$(_lets_active_command)
			_lets_command_options "${cmd}"
			;;
	esac
}

_check_lets_config() {
	${LETS_EXECUTABLE} "$@" 1>/dev/null 2>/dev/null
	echo $?
}

_lets_root_flags_before_command() {
	local idx=1
	local -a prefix=()

	while [ $idx -lt $CURRENT ]; do
		local token="${words[$idx]}"

		case "$token" in
			-c|--config|-E|--env|--only|--exclude)
				prefix+=("$token")
				((idx++))
				if [ $idx -lt $CURRENT ]; then
					prefix+=("${words[$idx]}")
				fi
				;;
			--config=*|--env=*|--only=*|--exclude=*|-E*)
				prefix+=("$token")
				;;
			-d|-dd|--debug|--all|--init)
				prefix+=("$token")
				;;
			--)
				break
				;;
			-*)
				prefix+=("$token")
				;;
			*)
				break
				;;
		esac

		((idx++))
	done

	reply=("${prefix[@]}")
}

_lets_active_command() {
	local idx=1

	while [ $idx -le $#words ]; do
		local token="${words[$idx]}"

		case "$token" in
			-c|--config|-E|--env|--only|--exclude)
				((idx++))
				;;
			--config=*|--env=*|--only=*|--exclude=*|-E*|-d|-dd|--debug|--all|--init|--)
				;;
			-*)
				;;
			*)
				echo "$token"
				return
				;;
		esac

		((idx++))
	done
}

_lets_commands () {
	local cmds
	_lets_root_flags_before_command
	local -a root_flags=("${reply[@]}")

	if [ $(_check_lets_config "${root_flags[@]}") -eq 0 ]; then
		IFS=$'\n' cmds=($(${LETS_EXECUTABLE} "${root_flags[@]}" completion --commands --verbose 2>/dev/null))
	else
		cmds=()
	fi
	_describe -t commands 'Available commands' cmds
}

_lets_command_options () {
	local cmd=$1
	_lets_root_flags_before_command
	local -a root_flags=("${reply[@]}")

	if [[ -z "$cmd" || "$cmd" == -* ]]; then
		return 0
	fi

	if [ $(_check_lets_config "${root_flags[@]}") -eq 0 ]; then
		IFS=$'\n'
		_arguments -s $(${LETS_EXECUTABLE} "${root_flags[@]}" completion --options=${cmd} --verbose 2>/dev/null)
	fi
}

if ! command -v compinit >/dev/null; then
	autoload -U compinit && compinit
fi

compdef _lets lets
`

const bashCompletionText = `_lets_completion() {
    cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=( $(lets completion --list "${COMP_WORDS[@]:1:$((COMP_CWORD-1))}" -- ${cur} 2>/dev/null) )
    if [[ ${COMPREPLY} == "" ]]; then
        COMPREPLY=( $(compgen -f -- ${cur}) )
    fi
    return 0
}

complete -o filenames -F _lets_completion lets
`

// generate bash completion script.
func genBashCompletion(out io.Writer) error {
	tmpl, err := template.New("Main").Parse(bashCompletionText)
	if err != nil {
		return fmt.Errorf("error creating zsh completion template: %w", err)
	}

	return tmpl.Execute(out, nil)
}

// generate zsh completion script.
func genZshCompletion(out io.Writer) error {
	tmpl, err := template.New("Main").Parse(zshCompletionText)
	if err != nil {
		return fmt.Errorf("error creating zsh completion template: %w", err)
	}

	return tmpl.Execute(out, nil)
}

// generate string of commands joined with \n.
func getCommandsList(rootCmd *cobra.Command, out io.Writer, verbose bool) error {
	buf := new(bytes.Buffer)

	for _, cmd := range rootCmd.Commands() {
		if !cmd.Hidden && cmd.Name() != "help" {
			if verbose {
				descr := "No description for command " + cmd.Name()
				if cmd.Short != "" {
					descr = cmd.Short
					descr = strings.TrimSpace(descr)
				}

				fmt.Fprintf(buf, "%s:%s\n", cmd.Name(), descr)
			} else {
				buf.WriteString(cmd.Name() + "\n")
			}
		}
	}

	_, err := buf.WriteTo(out)
	if err != nil {
		return fmt.Errorf("can not generate commands list: %w", err)
	}

	return nil
}

type option struct {
	name string
	desc string
}

// generate string of command options joined with \n.
func getCommandOptions(command *config.Command, out io.Writer, verbose bool) error {
	if command.Docopts == "" {
		return nil
	}

	rawOpts, err := docopt.ParseOptions(command.Docopts, command.Name)
	if err != nil {
		return fmt.Errorf("can not parse docopts: %w", err)
	}

	var options []option

	for _, opt := range rawOpts {
		if strings.HasPrefix(opt.Name, "--") {
			options = append(options, option{name: opt.Name, desc: opt.Description})
		}
	}

	sort.SliceStable(options, func(i, j int) bool {
		return options[i].name < options[j].name
	})

	buf := new(bytes.Buffer)

	for _, option := range options {
		if verbose {
			desc := "No description for option " + option.name

			if option.desc != "" {
				desc = strings.TrimSpace(option.desc)
			}

			fmt.Fprintf(buf, "%[1]s[%s]\n", option.name, desc)
		} else {
			buf.WriteString(option.name + "\n")
		}
	}

	_, err = buf.WriteTo(out)
	if err != nil {
		return fmt.Errorf("can not generate command options list: %w", err)
	}

	return nil
}

// InitCompletionCmd intializes root 'completion' subcommand.
// config can be nil, so we adding more flags only when config is found.
// Returns reinit function which must be called when config is parsed.
func InitCompletionCmd(rootCmd *cobra.Command, cfg *config.Config) func(cfg *config.Config) {
	completionCmd := &cobra.Command{
		Use:     "completion",
		Hidden:  true,
		Short:   "Generates completion scripts for bash, zsh",
		GroupID: "internal",
		RunE: func(cmd *cobra.Command, args []string) error {
			shellType, err := cmd.Flags().GetString("shell")
			if err != nil {
				return fmt.Errorf("can not get flag 'shell': %w", err)
			}

			if cfg != nil {
				verbose, err := cmd.Flags().GetBool("verbose")
				if err != nil {
					return fmt.Errorf("can not get flag 'verbose': %w", err)
				}

				list, err := cmd.Flags().GetBool("list")
				if err != nil {
					return fmt.Errorf("can not get flag 'list': %w", err)
				}

				commands, err := cmd.Flags().GetBool("commands")
				if err != nil {
					return fmt.Errorf("can not get flag 'commands': %w", err)
				}

				if list {
					commands = true
				}

				optionsForCmd, err := cmd.Flags().GetString("options")
				if err != nil {
					return fmt.Errorf("can not get flag 'options': %w", err)
				}

				if optionsForCmd != "" {
					command, exists := cfg.Commands[optionsForCmd]
					if !exists {
						return fmt.Errorf("command %s not declared in config", optionsForCmd)
					}

					return getCommandOptions(command, cmd.OutOrStdout(), verbose)
				}

				if commands {
					return getCommandsList(rootCmd, cmd.OutOrStdout(), verbose)
				}
			}

			if shellType == "" {
				return cmd.Help()
			}

			switch shellType {
			case "bash":
				return genBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return genZshCompletion(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell type %q", shellType)
			}
		},
	}

	completionCmd.Flags().StringP("shell", "s", "", "The type of shell (bash or zsh)")

	if cfg != nil {
		completionCmd.Flags().Bool("list", false, "Show list of commands [deprecated, use --commands]")
		completionCmd.Flags().Bool("commands", false, "Show list of commands")
		completionCmd.Flags().String("options", "", "Show list of options for command")
		completionCmd.Flags().Bool("verbose", false, "Verbose list of commands or options (with description) (only for zsh)")
	}

	rootCmd.AddCommand(completionCmd)

	return func(cfg *config.Config) {
		rootCmd.RemoveCommand(completionCmd)
		InitCompletionCmd(rootCmd, cfg)
	}
}
