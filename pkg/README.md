# pkg Directory

The _pkg_ directory contains a set of directories that are Go packages. These
packages are a Go library that other applications can import to make use of.
The functionality in the Hypper CLI that is not UI based is contained in these
packages.

For implementing these Hypper packages, we have chosen for now to extend by
composition the ones provided by Helm, which allows us to keep the same
structure meanwhile extending and providing _pkg_ for consumption.
