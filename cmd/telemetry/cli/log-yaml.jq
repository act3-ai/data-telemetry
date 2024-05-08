#   "red": "[31m",
#   "green": "[32m",
#   "yellow": "[33m",
#   "blue": "[34m",
#   "darkgray": "[90m",
#   "disabled": "[30;100m", # Black on darkgray
#   "reset": "[0m"

# pipe the logs to "jq -r -f log-yaml.jq"

# from https://stackoverflow.com/questions/53315791/how-to-convert-a-json-response-into-yaml-in-bash/53330236#53330236

def handleMultilineString($level):
      reduce ([match("\n+"; "g")]                       # find groups of '\n'
              | sort_by(-.offset))[] as $match
             (.; .[0:$match.offset + $match.length] +
                 "\n\("  " * $level)" +               # add one extra '\n' for every group of '\n's. Add indention for each new line
                 .[$match.offset + $match.length:]);

def toYamlString:
    if type == "string" and test("\n+"; "g")
    then 
        "|\n\(.)" | sub("\n"; "\n  "; "g")
    else .
    end;

def yamlify:
    (objects | to_entries[] | (.value | type) as $type |
        if $type == "array" then
            "\u001b[34m\(.key)\u001b[0m:", (.value | yamlify)
        elif $type == "object" then
            "\u001b[34m\(.key)\u001b[0m:", "  \(.value | yamlify)"
        else
            "\u001b[34m\(.key)\u001b[0m: \u001b[32m\(.value | toYamlString)\u001b[0m"
        end
    )
    // (arrays | select(length > 0)[] | [yamlify] |
        "  - \(.[0])", "    \(.[1:][])"
    )
    // .
    ;

# TODO handle multi-line strings as proper YAML multi-line strings

"\u001b[33m\(.level | ascii_upcase) \(.caller) \(.msg) \u001b[0m", 
del(.ts, .level, .caller, .msg) | yamlify
