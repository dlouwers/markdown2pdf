# Inline Code Wrapping Test Cases

This document tests various inline code wrapping scenarios to ensure proper handling of long code spans.

## Short Code (No Wrapping Needed)

This paragraph contains `short_code` that fits comfortably on one line without any issues.

## Long Variable Names

Testing underscores: `this_is_a_very_long_variable_name_that_should_break_at_underscores_to_fit_within_margins`

Testing camelCase: `thisIsAVeryLongCamelCaseVariableNameThatHasNoNaturalBreakPointsAndMightNeedForcedBreaking`

## Filesystem Paths

Unix path: `/very/long/filesystem/path/to/some/deeply/nested/directory/structure/file.txt`

Windows path: `C:\Users\SomeUser\Documents\Projects\markdown2pdf\internal\renderer\code_wrap.go`

URL: `https://github.com/dlouwers/markdown2pdf/blob/main/internal/renderer/text.go`

## Function Calls

Simple function: `calculateDiscountPercentage(basePrice, discountRate, taxMultiplier)`

Chained methods: `user.getProfile().getSettings().getNotificationPreferences().isEmailEnabled()`

## Mixed Punctuation

Namespace resolution: `com.example.project.package.ClassName.methodName()`

Array access: `data[very_long_index_variable_name][another_long_index]`

## Edge Cases

No break points: `supercalifragilisticexpialidociousthisisanextremelylongwordwithnobreakpoints`

Only spaces: `function call with many arguments that are just words`

Starting with separator: `_private_variable_name_starting_with_underscore`

Ending with separator: `variable_name_ending_with_underscore_`

Multiple separators: `path//with///multiple////slashes`

## Real-World Examples

Git commands: `git commit -m "Add inline code wrapping functionality" --no-verify --author="Author Name <email@example.com>"`

Docker commands: `docker run --name container_name --env VARIABLE_NAME=value -v /host/path:/container/path image:tag`

Configuration paths: `application.server.database.connection.pool.maxConnections`

## Inline Code in Context

Testing inline code mixed with normal text: The function `processUserAuthenticationRequestWithOAuthProviderValidation()` handles authentication. The path `/usr/local/lib/python3.9/site-packages/module/submodule/file.py` contains the implementation.

Multiple inline codes in one sentence: Use `first_function_with_long_name()` followed by `second_function_with_long_name()` and then `third_function_with_long_name()` to complete the process.

## Code with Special Characters

Regex pattern: `^[a-zA-Z0-9_-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`

SQL query: `SELECT column1, column2, column3 FROM very_long_table_name WHERE condition = 'value'`

JSON path: `$.data.items[0].attributes.metadata.tags[*].name`
