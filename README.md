# Hindsite website generator

[![Go Report Card](https://goreportcard.com/badge/github.com/srackham/hindsite)](https://goreportcard.com/report/github.com/srackham/hindsite)


## Overview
hindsite is a [fast](https://srackham.github.io/hindsite/faq.html#how-fast-is-hindsite),
lightweight static website
generator. It builds static websites with optional document and tag indexes from
Markdown source documents (e.g. blogs posts, newsletters, articles).

The hindsite stand-alone executable includes a built-in development web server
with Live Reload, a file watcher and incremental rebuilds.

- [Documentation](https://srackham.github.io/hindsite)
- [Downloads](https://github.com/srackham/hindsite/releases)
- [Github repository](https://github.com/srackham/hindsite)


## Quick Start
Create a fully functional blog and newsletter website with just [two hindsite
commands](https://srackham.github.io/hindsite/#quick-start).

    mkdir mysite
    hindsite init mysite -builtin blog
    hindsite serve mysite -launch
