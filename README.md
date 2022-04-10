# Hindsite website generator

[![Go Report Card](https://goreportcard.com/badge/github.com/srackham/hindsite)](https://goreportcard.com/report/github.com/srackham/hindsite)


## Overview
Hindsite is a
[fast](https://srackham.github.io/hindsite/faq.html#how-fast-is-hindsite),
lightweight static website generator. It builds static websites with optional
document and tag indexes from [Markdown](https://en.wikipedia.org/wiki/Markdown)
and [Rimu](https://github.com/srackham/rimu) source documents.

The hindsite stand-alone executable includes a built-in development web server
with Live Reload, a file watcher and incremental rebuilds.

- [Documentation](https://srackham.github.io/hindsite)
- [Downloads](https://github.com/srackham/hindsite/releases)
- [Github repository](https://github.com/srackham/hindsite)


## Quick Start
Create a fully functional blog and newsletter website with just [two hindsite
commands](https://srackham.github.io/hindsite/#quick-start).

```
mkdir hindsite-hello
cd hindsite-hello
hindsite init -from hello
hindsite serve -launch
```