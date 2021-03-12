# Installing Hypper

## From source (Linux, macOS)

Building Hypper from source is a bit more work, but the best way to test the
latest (pre-release) Hypper version.

You must have a working Go environment (1.14+).

```terminal
$ git clone https://github.com/rancher-sandbox/hypper/hypper.git
$ cd hypper
$ make
```

If required, it will fetch the dependencies and cache them, and validate
configuration.
It will then compile `hypper`, and place it in `./bin/hypper`.


## From development builds

TBD.

## From Github releases

TBD.
