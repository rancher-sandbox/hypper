# Installing Hypper

This guide shows how to install the Hypper CLI. Hypper can be installed either from source, or from pre-built binary releases.


## From binary releases (Linux, macOS, Windows)

Every [release](https://github.com/rancher-sandbox/hypper/releases) of Hypper provides binaries for a variety of OSes. These binary versions can be manually downloaded and installed.

  - Download your [desired version](https://github.com/rancher-sandbox/hypper/releases)
  - Unpack it (`tar -zxvf hypper-v0.0.2-linux-amd64.tar.gz`)
  - Find the hypper binary in the unpacked directory, and move it to its desired destination (`mv linux-amd64/hypper /usr/local/bin/hypper`)

From there, you should be able to run the client and add the stable repo: see `hypper help`.


## From script (Linux, macOS, windows)

Hypper now has an installer script that will automatically grab the latest version of hypper and install it locally.

You can fetch that script, and then execute it locally. It's well documented so that you can read through it and understand what it is doing before you run it.

```
$ curl -fsSL -o get_hypper https://raw.githubusercontent.com/rancher-sandbox/hypper/main/scripts/get-hypper
$ chmod 700 get_hypper
$ ./get-hypper
```

Yes, you can `curl https://raw.githubusercontent.com/rancher-sandbox/hypper/main/scripts/get-hypper | bash` if you want to live on the edge.


**NOTE**

*For installing using the script you need bash, curl/wget, and tar (openssl and sudo are optional).
Please check how you can install those requisites in your OS if needed.* 


## From source (Linux, macOS, Windows)

Building Hypper from source is a bit more work, but the best way to test the
latest (pre-release) Hypper version.

You must have a working Go environment (1.14+), and a POSIX environment with the proper tools installed (git, make).

For Windows, Hypper has been tested under [cygwin](https://www.cygwin.com/) and it compiles correctly.

```terminal
$ git clone https://github.com/rancher-sandbox/hypper/hypper.git
$ cd hypper
$ make
```

If required, it will fetch the dependencies and cache them, and validate
configuration.
It will then compile `hypper`, and place it in `bin/hypper` (`bin/hypper.exe` for windows).


## From development builds

In addition to releases you can download and install development snapshots of
Hypper from [Github Actions].

On each successful CI run, artifacts are stored for that run. Just download the binaries and follow the same steps as installing from binary releases.

**NOTE**

*Development snapshots could be in a broken state, with missing functionality or breaking changes.
We do not recommend using development snapshots on your production environment.*


## Conclusion

In most cases, installation is as simple as getting a pre-built hypper binary. This document covers additional cases for those who want to do more sophisticated things with Hypper.

Once you have the Hypper Client successfully installed, you can move on to the [quickstart guide](https://github.com/rancher-sandbox/hypper/docs/user/tutorials/quickstart.md).

[Github Actions]: https://github.com/rancher-sandbox/hypper/actions/workflows/ci.yml?query=branch%3Amain+is%3Asuccess+workflow%3ACI
