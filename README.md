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

## Quick Start
Install 
Create a fully functional blog and newsletter website with just [two hindsite
commands](https://srackham.github.io/hindsite/#quick-start).

```
mkdir hindsite-hello
cd hindsite-hello
hindsite init -from hello
hindsite serve -launch
```


## Quick Start
Create a fully functional blog and newsletter website with just two hindsite
commands:

1. Download the binary for your platform from the [releases
   page](https://github.com/srackham/hindsite/releases) and extract the
   `hindsite` executable (see [_Building hindsite_](https://srackham.github.io/hindsite/#building-hindsite) if you
   prefer to build and install from source).

2. Create a new Hindsite site directory and install the [built-in `blog`
   template]({hindsite-github}/tree/master/site/builtin/blog/template)
   using the hindsite [init command](#init-command):

    mkdir mysite
    hindsite init mysite -from blog

3. Build the website, start the hindsite web server and open it in the browser:

    hindsite serve mysite -launch

Include the `-v` (verbose) command option to see what's going on. The best way
to start learning hindsite is to browse the blog [site directory](#sites).

If you are familiar with Jekyll or Hugo you'll be right a home (hindsite can
read Jekyll and Hugo Markdown document [front matter](#front-matter)).

## Learn more
- [Documentation](https://srackham.github.io/hindsite)
- [Change log](https://srackham.github.io/hindsite/changelog.html)
- [Downloads](https://github.com/srackham/hindsite/releases)
- [Github repository](https://github.com/srackham/hindsite)