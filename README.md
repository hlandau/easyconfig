easyconfig: Easy bindings for configurable
==========================================

[![GoDoc](https://godoc.org/gopkg.in/hlandau/easyconfig.v1?status.svg)](https://godoc.org/gopkg.in/hlandau/easyconfig.v1)

easyconfig provides utilities for use with
[configurable](https://github.com/hlandau/configurable). It makes it easy to
get started with configurable using familiar interfaces.

Import as: `gopkg.in/hlandau/easyconfig.v1`

Packages provided
-----------------

The `cflag` package allows you to easily declare configuration items in a flag-like manner.

The `cstruct` package allows you to automatically generate configuration items from an annotated structure.

The `adaptflag` package adapts declared configuration items to flags and
registers them with the standard flag package and the
[pflag](https://github.com/ogier/pflag) package. You can also use it with any
flag package you like if it implements a similar registration interface.

The `easyconfig` package itself provides a simple struct-based configuration
interface; see the documentation in the examples.

Licence
-------

    © 2015 Hugo Landau <hlandau@devever.net>    MIT License

[Licenced under the licence with SHA256 hash
`fd80a26fbb3f644af1fa994134446702932968519797227e07a1368dea80f0bc`, a copy of
which can be found
here.](https://raw.githubusercontent.com/hlandau/acme/master/_doc/COPYING.MIT)
