# catjs
"catjs" is a versatile command-line utility designed for bug bounty hunters and security researchers. This tool specializes in JavaScript analysis, offering the unique ability to extract hidden endpoints and secrets from JavaScript code using JSON-based pattern detection. What sets "catjs" apart is its customizability – users can easily edit the "secret.json" file to define their own patterns, tailoring the tool to their specific requirements. Whether you're engaged in web security assessments, bug bounty programs, or penetration testing, "catjs" is your indispensable companion for identifying potential vulnerabilities and concealed security issues within web applications.

# Key Features:

1. Efficiently search for predefined patterns within JavaScript code using JSON-based detection.
2. Uncover hidden endpoints in JavaScript.
3. Discover and extract secrets concealed within JavaScript.
4. Customizable through the "secret.json" file, allowing users to define their own patterns.
5. Automatically save hidden endpoints to a file named `js_endpoint`.
6. Automatically save secret values to a file named `js_secret`.


# Installation

```
go install github.com/Ractiurd/catjs@latest
```

# Usage:

To analyze a single URL:
```
catjs -u <URL>
```

To analyze a list of URLs from a file:
```
catjs -f <file_path>
```

To read URLs from standard input (useful for piping data):
```
echo <URL> | catjs
```

Also user can use -c and -v for verbose and colored output

```
File location for secret.json must be Homedir/.config/secret.json
```

# Disclaimer:

Catjs is a tool designed for legitimate security research and bug bounty hunting. Ensure that you have proper authorization and adhere to responsible disclosure policies before using it on any target. so use it responsibly and ethically.

# Question:
If you have an question you can create an Issue or ping me on [Ractiurd](https://twitter.com/ractiurd)
