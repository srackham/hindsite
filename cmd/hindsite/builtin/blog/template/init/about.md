---
title: Hindsite built-in blog template
---

## Purpose
A Hindsite built-in template to create a website for blog posts and newsletters.

## Installation
To create, build and launch a new blog site run the following commands:

```
mkdir hindsite-blog
cd hindsite-blog
hindsite init -builtin blog
hindsite serve -launch
```

## Implementation
- The global site [configuration file](https://srackham.github.io/hindsite/index.html#configuration) is `template/config.toml`.
- The directory layout is documented in the [Hindsite Reference](https://srackham.github.io/hindsite/index.html#sites).
- The example content pages are written in Markdown but you could also [Rimu markup](https://github.com/srackham/rimu).
- There is an example _Search_ page that leverages Google programable searches.