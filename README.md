# Hindsite website generator

[![Go Report Card](https://goreportcard.com/badge/github.com/srackham/hindsite)](https://goreportcard.com/report/github.com/srackham/hindsite)
[![cover.run](https://cover.run/go/github.com/srackham/hindsite.svg?style=flat&tag=golang-1.10)](https://cover.run/go?tag=golang-1.10&repo=github.com%2Fsrackham%2Fhindsite)


## Overview
hindsite is a [very
fast](https://srackham.github.io/hindsite/faq.html#how-fast-is-hindsite),
lightweight static website generator. It builds static websites with optional
document and tag indexes from Markdown source documents (e.g. blogs posts,
newsletters, articles). Features:

- [High performance](https://srackham.github.io/hindsite/faq.html#how-fast-is-hindsite).
- Stand-alone single-file executable.
- Built-in development web server with Live Reload, file watcher and incremental
  rebuilds.


## Quick Start
Create a fully functional blog and newsletter website with just [two hindsite
commands](https://srackham.github.io/hindsite/#quick-start).

    mkdir myproject
    hindsite init myproject -builtin blog
    hindsite serve myproject -launch

## Links
Documentation: https://srackham.github.io/hindsite

Download page: https://github.com/srackham/hindsite/releases

Github: https://github.com/srackham/hindsite