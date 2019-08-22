# Project Scripts

The `scripts` directory is used for storing project related scripts.

## `godoc.sh`

Use `godoc.sh` to see the GoDoc of the project.  
At the moment, the built-in `godoc` (or `go doc`) tool does not support Go modules.
This script is a hack to test the GoDoc of this module. It uses docker internally.  
  
**Usage**  
From the project's root, run:
```
./scripts/godoc.sh
```
It prints out the URL to copy and paste in the browser. For example:
```
http://localhost:6060/pkg/github.com/jfrog/go-mockhttp
```