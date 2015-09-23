easyconfig: Easy bindings for configurable
==========================================

[![GoDoc](https://godoc.org/gopkg.in/hlandau/easyconfig.v0?status.svg)](https://godoc.org/gopkg.in/hlandau/easyconfig.v0)

easyconfig provides utilities for use with
[configurable](https://github.com/hlandau/configurable). It makes it easy to
get started with configurable using familiar interfaces.

Packages provided
-----------------

The `cflag` package allows you to easily declare configuration items in a flag-like manner.

The `cstruct` package allows you to automatically generate configuration items from an annotated structure.

The `adaptflag` package adapts declared configuration items to flags and
registers them with the standard flag package and the
[pflag](https://github.com/ogier/pflag) package. You can also use it with any
flag package you like if it implements a similar registration interface.

Licence
-------

    © 2015 Hugo Landau <hlandau@devever.net>    MIT License

