# Hindsite built-in `docs` template

## Purpose
A Hindsite template to create sites that have a fixed, relatively small, number
of pages. Ideal for building documentation sites e.g. User Guides, FAQs and
Reference Manuals.

This template was used to build the [Hindsite documentation site](https://srackham.github.io/hindsite/).

## Implementation
- The template features dynamic table of contents (TOC) generation (performed by
  JavaScript code in `main.js`).
- The header template includes the [highlight.js](https://highlightjs.org/)
  source code highlighter.
- The example content pages are written in Rimu markup. Rimu generates the
  header IDs which are used for table of contents generation (you could use
  Markdown instead of Rimu, but would need to modify the TOC generator code).

## Usage
To create, build and launch a new blog site run the following commands:

```
mkdir hindsite-docs
cd hindsite-docs
hindsite init -builtin docs
hindsite serve -launch
```