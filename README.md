# MMake (Mono Make)
## Introduction
Mono Make is a simple monorepo task runner based on top of Make. It allows you to split your Makefiles up across multiple files, and provides a few extra features to make your life easier when working with scripts in monorepos.

The goals here are to:
- Make it easy to split your Makefiles up across multiple files and directories
- Make it easy to discover and run targets
- Make it easy to run scripts from anywhere in your monorepo
- Make it easy to manage build outputs & code generation

This is still a work in progress, so there are a few features that are missing.

## Motivation
Make is a great tool for building software, but it has a few limitations when working with monorepos. The main limitation it is arduous to deal with multiple Makefiles. This means that if you want to split your Makefile up across multiple files, you need to use `include` statements or use -f to switch between files. This can get messy very quickly, and it can be hard to keep track of which Makefiles are included where.

MMake provides a simple solution to this problem. It allows you to split your Makefiles up across multiple files and directories, and it will automatically include them for you, inject variables that point to common directories for build outputs. It even provide autocomplete for targets.

If you're just starting out with a monorepo and want a unified tool to handle all your scripts this is probably a good place to start. If you're already using Make and want to split your Makefiles up across multiple files, this is also a good place to start.

## Installation
Requirements:
- Go 1.18 or higher

### Mac
```bash
go install github.com/aakarim/mmake
mmake init # This will add a WORKSPACE.mmake file
```

### Linux or WSL
```bash
go install github.com/aakarim/mmake
mmake init # This will add a WORKSPACE.mmake file
```

## Usage
Mono Make is a drop-in replacement for Make, so you can use it exactly the same way you would use Make and the files are just plain old Makefiles. The only difference is that you need to use `mmake` instead of `make` and you'll need to specify your targets using the root path syntax `//`.

To specify a target you can put an ordinary Makefile into any subdirectory of the root directory of your monorepo. This file will then be run from the context of the root directory of your monorepo. Environment variables will be injected into every file.

To specify a target you use the path to the target relative to your root directory specified by //, e.g. `//services/api:deploy`. This will run the `deploy` target in the `services/api/Makefile` file.

This is probably too simplistic for extremely complicated Makefiles. However, if you are running many services that have fairly simple and common tasks like build/test/deploy, then this should be more than enough. Also, it's a great way to throw together random scripts and discover them easily through the command line. 

### Clean
```bash
mmake clean //services/api
```
Will delete all the contents of the ./build-out/services/api directory.

### Autocompletion
Run the following command to enable autocompletion for Mono Make:
```bash
# This will be removed when your terminal resets. Add this command to your .bashrc or .zshrc to enable persistent autocompletion.
source <(mmake completion)

```

## Features
### Automatic Makefile inclusion
Mono Make will automatically include any Makefiles in the current directory and any child directories. This means you can split your Makefiles up across multiple files and directories.

### Environment Variable Injection
Mono Make will automatically inject environment variables into your Makefiles. It will inject the following variables:
- `MM_ROOT` - The root directory of the monorepo
- `MM_PATH` - The path to the current directory
- `MM_OUT_ROOT` - The path to the build output directory
- `MM_OUT_PATH` - The path to the build output directory for the 
current target

## Managed build outputs
Mono Make will automatically manage build outputs for you. It will create a `build-out` directory in the root directory of your monorepo. This directory will contain a directory for each target. When you run a target, it will create a directory for that target if it doesn't already exist. It will then run the target in that directory. This means that you can run the same target multiple times without having to worry about cleaning up the build output.

You should add `build-out` to your `.gitignore` file.

### Run from anywhere
You can run Mono Make from anywhere in your monorepo. It will automatically find the root directory of your monorepo and run the target from there.

### Target discovery & autocomplete
Mono Make will automatically discover targets in your Makefiles. It also provides autocomplete for targets. This means you can run `mmake //services/api:de` and it will autocomplete to `mmake //services/api:deploy` when you hit tab.

Target discovery also works through Makefile comments. If you have a kitchen-sink script that performs a few utilities you can use the command line autocomplete to find it easily. 

## Examples
### Simple Makefile
```makefile
default:
    @echo "Hello World"
```
```bash
mmake //services/api:default
# Output: Hello World
```

### Makefile with environment variables
```makefile
default:
	@echo "Root directory: $(MM_ROOT)"
	@echo "Current directory: $(MM_PATH)"
	@echo "Build directory: $(MM_OUT_ROOT)"
	@echo "Target directory: $(MM_OUT_PATH)"
```
```bash
mmake //services/api:default
# Output:
# Root directory: /Users/aakarim/Projects/monorepo
# Current directory: /Users/aakarim/Projects/monorepo/services/api
# Build directory: /Users/aakarim/Projects/monorepo/build-out
# Target directory: /Users/aakarim/Projects/monorepo/build-out/services/api
```

## Contributing
Contributions are very welcome!

## License
Mono Make is licensed under the MIT License. See [LICENSE](LICENSE) for more information.
```