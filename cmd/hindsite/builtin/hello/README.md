# Hindsite built-in _hello_ template

## Purpose
A minimal single-page Hindsite site.

## Structure
- A single `CONTENT_DIR/index.md` Markdown content document.
- A single `TEMPLATE_DIR/layout.html` HTML template.

## Usage
To create and launch a site based on this template run the following commands:

```
mkdir hello-hindsite
cd hello-hindsite
hindsite init -builtin hello
hindsite serve -launch
```