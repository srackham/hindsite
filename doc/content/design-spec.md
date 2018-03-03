# Hindsite Static Website Generator

Notes for yet another static website generator build with Go.

**Q**: Why?

**A**: I decided to start blogging again after a 3 year hiatis. I had been using
Hugo but my old blog would not build with the latest version of Hugo. Time to
debug! At first things when well, after renaming couple of template variable
identifiers my site compiled without errors, but I soon discovered that the
posts and categeories indexes were missing. I dug deeper and all the while was
nagged by the thought that for a simple blog the Hugo edifice was just way to
complex. Armed with typical programmer hubris I fancied I could build my own
generator, one that would be a doodle for mere mortals to learn and to use (and
reuse). Good luck with that!


# Names
hindsite (backup name hiatis).
Github name `hindsite`.


## Inspiration and resources
- [Writing a Static Blog Generator in Go](https://zupzup.org/static-blog-generator-go/).
- [How to code a markdown blogging system in Go](http://www.will3942.com/creating-blog-go).
- [Jekyll](https://jekyllrb.com).
- [Hugo](https://gohugo.io/).
- [Introduction to Hugo templating](https://gohugo.io/templates/introduction/)
- [Disqus comments](https://gohugo.io/content-management/comments/).


## Domain
Small static websites with optional indexed documents (blogs posts, newsletters, articles).


## Features
- Easy to get up and running with built-in scaffolding and layouts: install ->
  init -> build -> serve.
- Supports multiple groups of indexed documents each with separate tag and date
  indexes e.g. blog posts and newsletters.


## Simplest hindsite project

```
/content
    index.md
/template
    layout.html
/build
    index.html
```

index.md:
```
Hello *World*!
```

layout.html:
```
<!doctype html>
<head><title>.Title</title></head>
<body>.Body</body>
```

Content and template directories can be the same, not recommended but useful for
building an _example_.


## Design
The overarching goals are simplicity and ease of use.

- Hindsite textual content (Markdown files) is rendered with templates to
  generate a static website.
- Structural transparency: The content, template and build directory structures
  match -- content and static resouces reside in the same relative locations.
- Structural flexibility: There are few restrictions on directory structure and
  file locations.

- Just tags: no taxonomies, no categories see [WordPress Categories vs Tags: How
  Do I Use Them on My
  Blog?](https://www.wpsitecare.com/wordpress-categories-vs-tags/).
- No explicit slugs or permalinks.
- Minimal explicit meta-data.
- No _previous/next_ paging, just a _recent_, _all_, _tag_ and _tags_ indexes.
  Paging is just unnecessary complexity -- you can page indexes up and down and search them it in the browser.
- No [shortcodes](https://gohugo.io/content-management/shortcodes/) -- use Rimu macros.
- No template functions just a set of template variables, see:
  * https://jekyllrb.com/docs/variables/,
  * https://gohugo.io/variables/).
- No RSS -- hardly anyone even knows what it is anymore!


## Validation and testing
Here are a few of the many available tools.

### Local web server
In lieu of a built-in `serve` command (not yet implemented).

Run:

   python -m SimpleHTTPServer

Then open you web browser at http://localhost:8000/

I have the `ws` alias for this command.

### Link validation
Use the NPM [broken-link-checker](https://www.npmjs.com/package/broken-link-checker).

  blc -r -e http://localhost:8000/      # Recursively check all local links.
  blc -r http://localhost:8000/         # Recursively check all links.

### HTML validation
Use the NPM [html-validator-cli](https://www.npmjs.com/package/html-validator-cli).

```
  # Single document
  html-validator --verbose --format=text --file=doc.html
  # All .html docs recursively.
  for f in $(find . -name '*.html'); do
    echo $f
    html-validator --verbose --format=text --file=$f
  done
```

I added the `html-validator-all` alias to recursively validate all HTML files in current directory.

```
alias html-validator-all='for f in $(find . -name "*.html"); do echo $f; html-validator --verbose --format=text --file=$f; done'
```


## Coding
- Use the Go [html package](https://godoc.org/golang.org/x/net/html) to parse HTML.


## User notes
- Use root-relative links in templates because templates typically generate multiple rendered files
  at differing locations.
  See https://www.w3schools.com/html/html_filepaths.asp


## Vocabulary
project:: A hindsite project consists of a _content directory_, a _template
directory_ and a _build directory_. Hindsite uses the contents of the template
and content directories to generate a website which it writes to the build
directory.

content directory:: A directory (default name `content`) containing content
documents, optional document meta-data and optional static resources (e.g.
images). The content directory structure mirrors that of the template directory.

template directory:: A directory (default name `template`) containing webpage
templates along with optional configuration files and static resources to build
a website. The template directory is a blueprint, it contains everything needed
to build a Website except textual content.

build directory:: A directory (default name `build`) into which the built
Website is written.

content document:: A readable-text markup file (`.{md,rmu}`) that generates the
textual content of a corresponding HTML web page.

template:: A Go HTML template file (`.html`). Templates reside in the template
directory.

configuration file:: An optional TOML file containing site configuration
parameters called `config.toml` residing at the root of the template directory.

layout template:: Template files named `layout.html` used to render content
documents. A content document is rendered by first obtaining the list of layout
templates residing on the corresponding document path in the template directory.
The document is converted to HTML, passed to the first layout template and
rendered. Additional layout templates are passed the output of the previous
layout and rendered until all layouts have been processed.

index:: A generated HTML Webpage containing a list of content documents.

indexed directory:: A directory in the content directory whose corresponding
template directory contains one or more index templates
(`{all,recent,tags,tag}.html`).

indexed document:: A content document residing in an indexed directory. Indexed
documents must have a title and a publication date. Tags are optional.

date indexes:: hindsite generates date index pages ordered by document
publication date using `all.html` and `recent.html` templates. `all.html`
generates a list of all documents; `recent.html` generates a list of the most
recent documents.

tag:: A keyword or phrase used to categorise a content document.

tag indexes:: hindsite generates tag index pages using tag meta-data and
`tags.html` and `tag.html` templates.

static file:: A content file that is neither a document or a template. Static
files (typically HTML, CSS and images) are copied verbatim to the corresponding
location in the build directory.

example document:: A text markup file residing in the template director. Example
documents are are copied to the content directory when the project is
initialized. This provides a mechanism for sharing your website design complete
with example documents: just share your website template directory.


## Implementation
- By convention the content, template and build directories normally reside in
  the same folder, but this convention is not enforced.

- There is a close one-to-one mapping between content and build file paths. This
  lightens the congative load especially when it comes to authoring cross-page
  URLs.

Website with blog and newsletters:

```
/content
    config.rmu
    /pages
        index.md
        about.md
    /blog
        example.md
        foobar-image.png
        2008-12-10-post-zero.rmu
    /newsletters
        2018-02-07.md
        2018-02-14.md
    /static
/template
    config.toml
    layout.html
    CNAME
    favicon.ico
    /pages
        layout.html
    /blog
        layout.html
        all.html
        recent.html
        tags.html
        tag.html
        example.md
    /newsletters
        layout.html
        all.html
        recent.html
    /static
        main.css
        logo.jpg
/build
    favicon.ico
    CNAME
    /pages
        index.html
        about.html
    /blog
        post-one.html
        foobar-image.png
        2008-12-10-post-zero.html
    /newsletters
        2018-02-07.html
        2018-02-14.html
    /static
        main.css
        logo.jpg
    /indexes
        /blog
            all.html
            recent.html
            tags.html
            /tags
                kotlin.html
        /newsletters
            all.html
            recent.html
```

- Template files are preprocessed by the
  Go template engine.

/*
  If a same-named `.toml` is present it will be injected into the template.
  A `.toml` file with the same name as a directory is injected into all templates in the
  directory e.g. `blog.toml` will be injeced into all documents in the `blog` directory.
*/

- Reserved directory names: `/indexes`. By convention the `/static`
  directory contains static asset files, but this is not enforced.

- The `/indexes` build directory contains a sub-directory for each indexed directory.
  Each sub-directory may contain:

  * `recent.html`: a list of the most recent posts.
  * `all.html`: a list of the all posts.
  * `tags.html`: a list of post tags linked to files named like `tags/TAG_NAME.html` (one per tag)
    each containing a list of related posts. The `TAG_NAME` is slugified.


## File and directory names
Content file and directory names should
only contain lower case alpha numeric and hyphen characters. This is is mandatory.
and ensures sane readable URLs without ugly encodings for illegal URL characters.

If necessary use the `-slugify` option will nomalize all content paths.


## Template variables
Template variables contain content, content meta-data, configuration data along
with context-specific data.

Variable contexts:

- Content document.
- Date index templates.
- Tag index templates.

/*
Configuration data and document meta-data Meta-data is aggregated before for
template substitution as follows:

1. Global `config.toml` (in root of content directory).
2. Local `config.toml` (in template directory).
3. Document meta-data.

Meta-data priority: document (highest), local, global (lowest).
*/

Example global `config.toml` file:

```
# Is prepended to all root-relative URLs.
# DO NOT IMPLEMENT UNTIL NEEDED.
rootPrefix= "/stuarts-notes"

# Use this index file if there is no /index.html file.
homePage = "/indexes/blog/recent.html"

# Maximum number of recent index entries.
maxRecent = 100

# Default document author.
author = "Joe Bloggs"

# date/time.
timezone = "Local"
shortDateFormat = ""
mediumDateFormat = ""
longDateFormat = ""
```

The optional `config.rmu` file in the content directory will be prepended to all Rimu files
before they are rendered and mostly used to define macro values.

Per-file meta-data can be sourced from:

/*
- A same-named `.toml` file.
*/
- Embedded parameters at the start of content files (c.f.
  https://www.npmjs.com/package/marked-metadata,
  https://gohugo.io/content-management/front-matter/):

Embedded YAML with `<!--`, `-->` or `---` delimiters e.g.

```
<!--
title:    Foo Bar
synopsis: An brief history of the origins of the foofoo valve.
addendum: This article was updated March, 2018.
date:     2018-02-16
tags:     foo foo valve, qux, baz
draft:    true
-->
```
- `tags` can be a comma-separated string or a YAML list:
- The HTML comment delimiters are preferable because:
  * They ignored if the document is processed by other applications.
  * The front matter is retained by the generated HTML web page.
- [YAML syntax](https://learn.getgrav.org/advanced/yaml).
- See [go-yaml](https://github.com/go-yaml/yaml) package.

Embedded TOML (c.f. Hugo front matter):

```
+++
title = "Foo Bar"
description = "An brief history of the origins of the foofoo valve."
date = "2018-02-16"
categories = [
  "foo foo valve",
  "qux",
  "baz"
]
draft = "true"
+++
```
Key aliases to allow Hugo front matter consumption: description = synopsis; categories = tags.

- Implicit values (if the document does not contain a front matter header):
  * `title` defaults to first `h1` title or, failing that, the document file
    name (sans drafts and date prefixes with hyphens replaced by spaces).
  * `date` defaults to file name `YYYY-MM-DD` date prefix e.g.
    `2018-02-14-newsletter.md`.
  * `author` defaults to value in `config.toml` or, failing that, blank.
  * `synopsis` defaults to the first paragraph.
  * `addendum` defaults to blank.
  * `draft` is true if the first letter of the content file name is a tilda.

If you don't use tags then you don't need explicit external or embedded meta-data.

## CLI
Usage (c.f. `go help`and `dlv help`): 

```
hindsite command [arguments]

The commands are:

    init
    build
    serve

Use "go help [command]" for more information about a command.

hindsite init [-project PROJECT_DIR] [-template TEMPLATE_DIR] [-content CONTENT_DIR]
hindsite build [-drafts] [-slugify] [-project PROJECT_DIR] [-template TEMPLATE_DIR] [-content CONTENT_DIR] [-build BUILD_DIR]
hindsite serve [-project PROJECT_DIR] [-build BUILD_DIR]
hindsite -h | --help | help

/*
hindsite check [CONTENT_DIR]
hindsite validate [BUILD_DIR]
hindsite slugify [-exclude FILE] [CONTENT_DIR]
*/
```
- Content, template and build directories must be mutually exclusive (not
  contained within one another) with one exception: content and template
  directories can be identical (this allows template examples to be tested).
- If content, template or build directory paths are relative they are relative
  to the project directory.
- `PROJECT_DIR` defaults to `.`
- `CONTENT_DIR` defaults to `content`
- `BUILD_DIR` defaults to `build`
- `TEMPLATE_DIR` defaults to `template`
- All commands support the `-v`,`-verbose` option.

### Init command
Creates new project content directory from an existing template directory.

- Clones the `TEMPLATE_DIR` directory structure to `CONTENT_DIR`.
- Copies example content documents to `CONTENT_DIR`.
- Creates `CONTENT_DIR` directory (if the `CONTENT_DIR` already exists init will
  refuse to continue).


/*
### Check command
??? But this is covered by the build command.

Checks and reports:

- Non-slugified document and directory names.
- Missing or invalid indexed document meta-data.
*/

### Build command
Build static website from content and template directories.

- Include draft pages if the `-drafts` option is specified.
- Slugify content document paths if the  `-sligify` option is specified.

Build sequence:

- Check content, template and build directories exist.
- Delete contents of build directory.
- Copy static files from content to build directory.
- Copy static files from template to build directory (do not overwrite existing
  files).
- Generate all non-indexed documents.
- Process each indexed directory generating indexes and document web pages.

/*
### Validate command
??? Shouldn't this choice be left to external tools such as blc?

Performs HTML validation of the built website.

- Run all HTML pages in `BUILD_DIR` through the W3C Nu validator.
*/

### Serve command
Open a development server on the `BUILD_DIR` directory.


/*
### Slugify command
??? But this is is covered by the build command `-slugify` option.

Normalize all file and directory path names in the content directory to
only contain lower case alpha numeric and hyphen characters.

- Multiple `-exclude FILE` options are allowed.
- The excluded file name is relative to the `CONTENT_DIR`.
*/

## Implementation plan
### Stage 1
- Start by implementing `init`, `build` and `help` commands.


## Implementation ideas
- Add Disqus comments -- I suspect this is just a content templating issue.


## Error handling
Use the Go [log](https://golang.org/pkg/log/) package.


## Dates
### Meta data
shortDate, mediumDate, longDate::
Publication date in the config file shortDateFormat, mediumDateFormat and longDateFormat formats.

Use maps for data, example

``` go
package main
import "fmt"
import "os"
import "html/template"

type TemplateData map[string]interface{}

func merge(m TemplateData) {
	m["qux"]=template.HTML("<baz>")
}

func main() {
        m := TemplateData{"foo":7,"bar":"<baz>"}
	      merge(m)
        t, err := template.New("test").Parse("{{.foo}} {{.bar}} {{.qux}}")
        if err != nil { fmt.Println(err); return }
        err = t.Execute(os.Stdout, m)	// Prints: 7 &lt;baz&gt; <baz>
        if err != nil { fmt.Println(err); return }
}
```

### Input formats
The RFC 3339 format is recognised. The following relaxations are permitted:

- If the time zone offet is omitted then the `timezone` configuration value is used;
  if `timezone` is not specified or set to `"Local"` then local time offset is used.
- If the time is omitted then the time `00:00:00` is used.
- A space character separator (instead of the letter `T`) between the date and the time is allowed.

e.g. with `timezone = "+13:00"`:

    1979-05-27T07:32:00+13:00   # follows the RFC 3339 spec
    1979-05-27T07:32:00         # 1979-05-27T07:32:00+13:00
    1979-05-27 07:32:00         # 1979-05-27T07:32:00+13:00
    1979-05-27                  # 1979-05-27T00:00:00+13:00

Parser:

``` go
date := "2018-02-27T14:45:00"
if timezone == nil || timezone == "Local" {
    loc, _ := time.LoadLocation("Local")
    t, err := time.ParseInLocation(time.RFC3339, date, loc)
} else {
    t, err := time.Parse(time.RFC3339, date+timezone)
}
```

See:

- https://www.ietf.org/rfc/rfc3339.txt
- https://golang.org/pkg/time/

