# Hindsite website generator

[![Go Report Card](https://goreportcard.com/badge/github.com/srackham/hindsite)](https://goreportcard.com/report/github.com/srackham/hindsite)


## Overview
Hindsite is a
[fast](https://srackham.github.io/hindsite/faq.html#how-fast-is-hindsite),
lightweight static website generator. It builds static websites with optional
document and tag indexes from [Markdown](https://en.wikipedia.org/wiki/Markdown)
and [Rimu](https://github.com/srackham/rimu) source documents.

The Hindsite stand-alone executable includes:

- Built-in site templates to get you up and running quickly.
- A development web server with live reload and incremental rebuilds.
- A linter for validating generated webpages.

## Quick Start
Create a fully functional blog and newsletter website with just two hindsite
commands:

    mkdir myblog
    cd myblog
    hindsite init -from blog
    hindsite serve -launch

## Learn more
- [Documentation](https://srackham.github.io/hindsite)
- [Change log](https://srackham.github.io/hindsite/changelog.html)
- [Github repository](https://github.com/srackham/hindsite)