# Hindsite built-in `hello` template

## Purpose
A minimal single-page Hindsite bultin template.

## Implementation
- A single `CONTENT_DIR/index.md` Markdown content document.
- A single `TEMPLATE_DIR/layout.html` HTML template.

## Usage
To create, build and launch a site based on this template run the following commands:

```
mkdir hindsite-hello
cd hindsite-hello
hindsite init -builtin hello
hindsite serve -launch
```