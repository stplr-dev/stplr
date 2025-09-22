# fish completion for stplr

function __stplr_perform_completion
    # Extract all args except the last one
    set -l args (commandline -opc)
    # Extract the last arg (partial input)
    set -l lastArg (commandline -ct)
    # Call stplr --generate-shell-completion
    set -l results ($args[1] $args[2..-1] $lastArg --generate-shell-completion 2> /dev/null)

    # Remove trailing empty lines
    for line in $results[-1..1]
        if test (string trim -- $line) = ""
            set results $results[1..-2]
        else
            break
        end
    end

    for line in $results
        if not string match -q -- "stplr*" $line
            set -l parts (string split -m 1 ":" -- "$line")
            if test (count $parts) -eq 2
                printf "%s\t%s\n" "$parts[1]" "$parts[2]"
            else
                printf "%s\n" "$line"
            end
        end
    end
end

# Clear existing completions for stplr
complete -c stplr -e

# Register completion function
complete -c stplr -f -a '(__stplr_perform_completion)'