# MMake (Mono Make): Monorepo Task Runner
    
## About
Mono Make is a simple monorepo task runner based on top of Make. It allows you to split your Makefiles up across multiple files, and provides a few extra features to make your life easier when working with scripts in monorepos.

The goals here are to:
- Simplify the distribution of Makefiles across various files and directories.
- Make it easy to discover and run targets.
- Enable scripts to run anywhere within your monorepo.
- Manage build outputs & code generation

## Motivation
Despite the capabilities of Make, managing multiple Makefiles in monorepos can be complicated. MMake helps you distribute Makefiles across multiple files and directories. It automatically includes these files, inserts variables pointing to common directories for build outputs, and provides autocomplete for targets.

If you're setting up a monorepo and need a unified script handling tool, MMake could be useful. It's also beneficial if you're using Make and looking to distribute your Makefiles across various files.

## Installation
Requirements:
- Go 1.18 or higher

### Mac & Linux/WSL
```bash
go install github.com/aakarim/mmake
mmake init # This will add a WORKSPACE.mmake file
```

### Autocompletion 
Run the following command to enable autocompletion for Mono Make:
```bash
# This will be removed when your terminal resets. Add this command to your .bashrc or .zshrc to enable persistent autocompletion.
source <(mmake completion)
```

```bash
mmake //services/api:de<TAB> # Autocompletes to mmake //services/api:deploy
```
## Features
### Automatic Makefile inclusion
MMake automatically includes Makefiles from the current and child directories.

### Environment Variable Injection
Mono Make will automatically inject environment variables into your Makefiles. It will inject the following variables:
- `MM_ROOT` - The root directory of the monorepo
- `MM_PATH` - The path to the current directory
- `MM_OUT_ROOT` - The path to the build output directory
- `MM_OUT_PATH` - The path to the build output directory for the 
current target

### Managed build outputs
MMake automatically creates and manages a build-out directory in the monorepo root directory for each target. 

The build-out directory should be added to your `.gitignore` file.

### Run from anywhere
MMake can be run from any location within your monorepo.

### Target discovery & autocomplete
MMake automatically discovers Makefile targets and provides autocomplete.

## Usage
MMake functions as a replacement for Make, allowing you to use it in the same manner as Make. It recognizes regular Makefiles, but you should use mmake instead of make and specify your targets using the root path syntax `//`.

To define a target, place a Makefile into any subdirectory of the monorepo root directory. This file will run in the context of the root directory. Environment variables will be injected into each file.

A target is specified using the path relative to the root directory marked by `//`, e.g. `//services/api:deploy`. This runs the deploy target in the `services/api/Makefile` file.

While MMake might not be suitable for very complex Makefiles, it's efficient for managing services with simple and common tasks like build/test/deploy. Also, it's a great way to throw together scripts and discover them easily through the command line. 

### Clean
```bash
mmake clean //services/api
```
Will delete all the contents of the `./build-out/services/api directory`.

### Query
# Need to change the current query to complete and make a new one

## Examples
Check the provided Makefile examples for an idea of how MMake operates.

## Contributing
Contributions are welcome!

## License
Mono Make is licensed under the MIT License. See [LICENSE](LICENSE) for more information.
```