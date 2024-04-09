# servus

> HTTP server with built-in automatic refresh feature on file changes for developers.

## usage

```sh
servus                # watches current directory (non-recursive)
servus `git ls-files` # watches all files in git repository
PORT=4269 servus      # set custom port number 
```

## install

```sh
go install github.com/balazs4/servus
```

or download [pre-built binary](https://github.com/balazs4/servus/releases).

## license

see [license](./license)

## author

balazs4
