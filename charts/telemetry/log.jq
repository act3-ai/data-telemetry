["\u001b[33m", "\u001b[31m", "\u001b[34m", "\u001b[90m", "\u001b[0m"] as [$yellow, $red, $blue, $grey, $stop] |
$yellow, (.level | ascii_upcase), " ", .msg, $stop, " ", .source.file, ":", .source.line, " ",
del(.time, .level, .source, .msg, .stacktrace, .contents),
if .stacktrace? == null then "" else $red, "\nstack trace follows:\n", .stacktrace, $stop, "\n" end,
if .sql? == null then "" else $grey, "\nSQL (unescaped): ", .sql, "\u001b[0m", "\n" end,
"\n"