# Core

Core creates CoreOS virtual machines on OS X using Xhyve.

## Prerequisites

- xhyve
- e2fsprogs (optional)

```
brew install xhyve
brew install e2fsprogs
```

## Creating a disk image

Run the following to create a 5G root disk image (requires e2fsprogs).

```
dd if=/dev/zero of=./xhyve.img bs=1m count=5000
/usr/local/opt/e2fsprogs/sbin/mkfs.ext4 -L ROOT xhyve.img
```

Then to use this as the root partition for the VM:

```
core run --root=xhyve.img
```
