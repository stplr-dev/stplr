# Adding Packages to LURE's repo

## Requirements

- `go` (1.18+)
- `git`
- `lure-analyzer`
    - `go install go.elara.ws/lure-repo-bot/cmd/lure-analyzer@latest`
- `shfmt`
    - May be available in distro repos
    - `go install mvdan.cc/sh/v3/cmd/shfmt@latest`

---

## How to test a package

To test packages you can first create [a `lure.sh` shell file](./build-scripts.md) and then run the `lure build` comand to build the local `lure.sh` file into a package for your distro (more info about the `build` command [here](./usage.md#build)). You can then install this file to your distro and test it.

## How to submit a package

LURE's repo is hosted on Github at https://github.com/Elara6331/lure-repo. In it, there are multiple directories each containing a `lure.sh` file. In order to add a package to LURE's repo, simply create a PR with a [build script](./build-scripts.md) and place it in a directory with the same name as the package.

Upon submitting the PR, [lure-repo-bot](https://github.com/Elara6331/lure-repo-bot) will pull your PR and analyze it, providing suggestions for fixes as review comments. If there are no problems, the bot will approve your changes. If there are issues, re-request review from the bot after you've finished applying the fixes and it will automatically review the PR again.

All scripts submitted to the LURE repo should be formatted with `shfmt`. If they are not properly formatted, Github Actions will add suggestions in the "Files Changed" tab of the PR.

Once your PR is merged, LURE will pull the changed repo and your package will be available for people to install.
