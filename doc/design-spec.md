# Hindsite Static Website Generator

Notes for yet another static website generator build with Go.

**Q**: Why?

**A**: I decided to start blogging again after a 3 year hiatus. I had been using
Hugo but my old blog would not build with the latest version of Hugo. Time to
debug! At first things when well, after renaming couple of template variable
identifiers my site compiled without errors, but I soon discovered that the
posts and categories indexes were missing. I dug deeper and all the while was
nagged by the thought that for a simple blog the Hugo edifice was just way to
complex. Armed with typical programmer hubris I fancied I could build my own
generator, one that would be a doodle for mere mortals to learn and to use (and
reuse). Good luck with that!


# Names
hindsite (backup name hiatus).
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
- Easy to get up and running with built-in layouts: install ->
  init -> build -> serve.
- Supports multiple categories of indexed documents each with separate tag and
  date indexes e.g. blog posts and newsletters.


## Hello hindsite
The simplest project.

`index.md` file:
```
Hello *World*!
```

`layout.html` file:
```
<!doctype html>
<head><title>{{.title}}</title></head>
<body>{{.body}}</body>
```

### Using the [Standard project model](#standard-project-model)
Build command: `hindsite build`

```
content/
    index.md
template/
    layout.html
build/
    index.html
```

### Using the [Minimal project model](#minimal-project-model)
Build command: `hindsite build -content . -template .`

```
index.md
layout.html
build/
    index.html
```


## Design
The overarching goals are simplicity and ease of use.

- Hindsite textual content (Markdown files) is rendered with templates to
  generate a static website.
- Structural transparency: The content, template and build directory structures
  match -- content and static resouces reside in the same relative locations.
  Same directory model for content, template and build directories.
- Structural flexibility: There are few restrictions on directory structure and
  file locations.

- Just tags: no taxonomies, no categories see [WordPress Categories vs Tags: How
  Do I Use Them on My
  Blog?](https://www.wpsitecare.com/wordpress-categories-vs-tags/).
- No explicit slugs or permalinks.
- Minimal explicit meta-data.
- No pagination, just complete _recent_, _all_, _tag_ and _tags_ indexes.
  Pagination is an unnecessary complexity -- you can page indexes up and down and search them it in the browser.
- No [shortcodes](https://gohugo.io/content-management/shortcodes/) -- use Rimu macros.
- No template functions just a set of template variables, see:
  * https://jekyllrb.com/docs/variables/,
  * https://gohugo.io/variables/).
- No RSS -- hardly anyone even knows what it is anymore!


## Template variables
Template variables are injected into template files during webpage generation.
There are template variables for document front matter, index generation,
configuration parameters and miscellaneous items such as the current date and
time.

Inside a template, variable names are prefixed with a dot and enclosed by double
curly braces. For example the document `title` variable appears in the template
as`{{.title}}`. See the Go [html/template
package](https://golang.org/pkg/html/template/) more about the use of templates.

Template variable names are:

- Alpha-numeric starting with an alpha character.
- Case sensitive.


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


## Vocabulary
**project**: A hindsite [project](#projects) consists of a _content directory_,
a _template directory_ and a _build directory_. Hindsite uses the contents of
the template and content directories to generate a website which it writes to
the build directory.

**content directory**: A directory (default name `content`) containing content
documents along with optional document meta-data (front matter) and optional
static resources (e.g. images). The content directory structure mirrors that of
the template directory.

**template directory**: A directory (default name `template`) containing webpage
templates along with optional configuration files and static resources to build
a website. The template directory is a blueprint, it contains everything needed
to build a Website minus the textual content.

**build directory**: A directory (default name `build`) into which the generated
Website is written.

**content document**: A readable-text markup file (`.{md,rmu}`) that generates the
textual content of a corresponding HTML web page.

**configuration file**: Optional TOML or YAML files `config.{toml,yaml}`
containing configuration parameters located in the content
template directories.

**template**: A Go HTML template file (`.html`). Templates reside in the template
directory. Templates are combined with document content and meta-data to
generate HTML website pages.

**document templates**: Document templates are used to render content documents.
Each content document has a template named `layout.html` in the corresponding
template directory. If there is no `layout.html` in the corresponding template
directory then the parent template directories are searched right up to the
templates root directory.

**index templates**: Index templates are used to render document index web pages.
They are named `{all,recent,tag,tags}.html`.

**index**: A generated HTML Webpage containing a list of content documents.

**indexed directory**: A directory in the content directory whose corresponding
template directory contains one or more index templates.

**indexed document**: A content document residing in an indexed directory. Indexed
documents must have a title and a publication date. Tags are optional.

**date indexes**: hindsite generates date index pages ordered by document
publication date using `all.html` and `recent.html` index templates. `all.html`
generates a list of all documents; `recent.html` generates a list of the most
recent documents.

**tag**: A keyword or phrase used to categorise a content document.

**tag indexes**: hindsite generates tag index pages using tag meta-data and
`tags.html` and `tag.html` templates.

**static file**: A file in the content or template directories that is not a
document or a template or a configuration file. Static files (typically CSS and
images) are copied verbatim to the corresponding location in the build directory
by the hindsite build command. Static files in the content directory take
precedence over static files in the template directory.

**example document**: A text markup document residing in the template directory.
Example documents are are copied to the content directory when the project is
initialized. This provides a mechanism for sharing your website design complete
with example documents: just share your website template directory.

**slugify**: Ensure path names only contain lower case alpha numeric and hyphen
characters.

**front matter**: Meta-data embedded at the start of content documents.


## Implementation
- By convention the content, template and build directories normally reside in
  the same folder, but this convention is not enforced.

- There is a close one-to-one mapping between content, template and build file
  paths. This lightens the congative load especially when it comes to authoring
  cross-page URLs.

Website with blog and newsletters:

```
content/
    config.toml
    config.rmu
    pages/
        index.md
        about.md
    blog/
        example.md
        foobar-image.png
        2008-12-10-post-zero.rmu
    newsletters/
        2018-02-07.md
        2018-02-14.md
    static/
template/
    config.toml
    layout.html
    CNAME
    favicon.ico
    pages/
        layout.html
    blog/
        layout.html
        all.html
        recent.html
        tags.html
        tag.html
        example.md
    newsletters/
        layout.html
        all.html
        recent.html
        example.md
    static/
        main.css
        logo.jpg
build/
    favicon.ico
    CNAME
    pages/
        index.html
        about.html
    blog/
        post-one.html
        foobar-image.png
        2008-12-10-post-zero.html
    newsletters/
        2018-02-07.html
        2018-02-14.html
    static/
        main.css
        logo.jpg
    indexes/
        blog/
            all.html
            recent.html
            tags.html
            tags/
                kotlin.html
        newsletters/
            all.html
            recent.html
```

- Template files are preprocessed by the
  Go template engine.

/*
  If a same-named `.toml` is present it will be injected into the template.
  A `.toml` file with the same name as a directory is injected into all templates in the
  directory e.g. `blog.toml` will be injected into all documents in the `blog` directory.
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
Content file and directory names should only contain lower case alpha numeric
and hyphen characters. This is is mandatory, it:

* Ensures consistent, readable URLs without ugly encodings for illegal URL
  characters.
* Allows file names to map to reasonable document titles.

If necessary use the build command `-slugify` option to nomalize all content paths.


## URLs
Hindsite synthesizes document URLs for index page links. By default synthesized
URLs are root-relative to the content directory. For example:

    Content directory: /tmp/project/content
    Content document:  /tmp/project/content/posts/post1.md
    Synthesised URL:   /posts/post1.html

This works fine if you deploy the build to the root of your website. For example
if your website address is `http://example.com` then the synthesised
root-relative `/posts/post1.html` URL is equivalent to
`http://example.com/posts/post1.html`.

If you are deploying to a server subdirectory then you need to set the
`urlprefix` configuration variable to ensure synthesized URLs are rooted
correctly. For example, if are deploying to the server `http://example.com/blog`
directory you need to set `urlprefix` to `/blog`. The synthesized URL from the
previous example now becomes `/blog/posts/post1.html` (equivalent to
`http://example.com/blog/posts/post1.html`).

To synthesize absolute URLs include the URL domain in `urlprefix`. Using the
previous example, setting `urlprefix` to `http://example.com/blog` will
synthesize the absolute URL `http://example.com/blog/posts/post1.html`.

Use relative URLs for authored links between document pages -- it's less verbose
and side-steps root-relative URL issues. For example `./post2.html` links to a
document in the same directory.

See [HTML File Paths](https://www.w3schools.com/html/html_filepaths.asp).

If you want to hide URL file name extensions you will need to configure your
server, see [How to remove file extension from website
address?](https://stackoverflow.com/questions/6534904/how-to-remove-file-extension-from-website-address).


## Indexes
- Recent, All and Tags document indexes are created by the build command.
- The project can contain any number of indexed content directories.
- Any content directory can be indexed by placing one or more index
  templates in the corresponding template directory.
- An index includes all documents from all subdirectories.
- Indexes can be nested.
- By default all built index pages are written to the `indexes` directory
  located at the root of the build directory.
- The location of the built index pages can be changed to anywhere inside the
  build directory using the build command `-index INDEX_DIR` option.


## Template variables
Template variables contain content, content meta-data, configuration data along
with context-specific data.

Variable contexts:

- Layout (content document) templates.
- Date index templates.
- Tag index templates.

## Configuration and front matter variables
Additional build information is sourced from configuration files and document front matter.

### Configuration files
TODO: Should multiple configuration files in the document path be aggregated
(template path then content path, top to bottom)?

- Configuration files and front matter are optional.
- Configuration files and front matter are encoded in TOML or YAML (take your pick).
- Configuration files are named `config.{toml,yaml}` and are located and are
  located at the roots of the contents and templates directories (values from
  the latter take precedence).

Example `config.toml` file:

```
# prepended to all synthesised root-relative URLs.
urlprefix = "/blog"

# Site language code.
lang = "en"

# Build uses this to copy a file to /index.html (assuming you need it and it do not have it).
# The value is a file path (either absolute or relative to the build directory).
homePage = "indexes/blog/recent.html"

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

Use maps for template data, example

``` yaml
animals:
  horse: big
  mouse: small
```

``` toml
[animals]
horse = "big"
mouse = "small"
```

Generates a Go field `Animals` containing a map:

```
Animals: map[string]string{
  "horse": "big",
  "mouse": "small",
}
```

``` go
package main
import "fmt"
import "os"
import "html/template"

type TemplateData map[string]interface{}

func merge(m TemplateData) {
	m["qu_x"]=template.HTML("<baz>")
	m["baz"] = map[string]string{"a":"A","b":"B"}
}

func main() {
        m := TemplateData{"foo":7,"bar":"<baz>"}
	      merge(m)
        t, err := template.New("test").Parse("{{.foo}} {{.bar}} {{.qu_x}} {{.baz.a}} {{.baz.b}}")
        if err != nil { fmt.Println(err); return }
        err = t.Execute(os.Stdout, m)	// Prints: 7 &lt;baz&gt; <baz> A B
        if err != nil { fmt.Println(err); return }
}
```


## Front matter
Document front matter is embedded at the start of content files (c.f.
https://www.npmjs.com/package/marked-metadata,
https://gohugo.io/content-management/front-matter/).

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
- `tags` are a comma-separated string. Tags can also be encoded as a YAML list
  or a TOML array called `categories` c.f. Hugo.
- The variable name `description` can be used instead of `synopsis`.
- The HTML comment delimiters are preferable because:
  * They ignored if the document is rendered by other applications.
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

Implicit values (if the document does not contain a front matter header):

- `title` defaults the document file name (sans drafts and date prefixes with
  hyphens replaced by spaces). NOTE: decided not to extract first `h1` title as
  that necessitates rendering the page first.
- `date` defaults to file name `YYYY-MM-DD-` date prefix e.g.
  `2018-02-14-newsletter.md`.
- `author` defaults to the configuration file value, failing that, blank.
- `synopsis` defaults to blank. NOTE: decided not to extract the first paragraph
  as that necessitates rendering the page first.
- `addendum` defaults to blank.
- `draft` is true if the first letter of the content file name is a tilda (this
  overrides the front matter `drafts` variable).


## Projects
A hindsite project consists of a _content directory_, a _template
directory_ and a _build directory_. Hindsite uses the contents of the template
and content directories to generate a website which it writes to the build
directory.

Content, template and build directories must be mutually exclusive (not
contained within one another) with the following execptions:

- Content and template directories can be identical.
- The build directory can be located at the root of the content directory.

These exceptions accomodate the [Minimal project model](#minimal-project_model).

Normally it's not a good idea to mix content and
  design especially if you plan to share or reuse your project template.

This allows template
  examples to be tested and may be a useful model for simple sites (the hindsite
  docs use this model).
But normally it's not a good idea to mix content and
  design especially if you plan to share or reuse your project template.

TODO: Document _Standard_, _Minimal_ and _Disjoint_ project models.

### Standard project model
Separate content, template and build directories in a single directory. Content
and design are separate. This is the default project model.

```
content/
    content files...
template/
    template files...
build/
    website files...
```
Standard project model build command: `hindsite build`

### Disjoint project model
Same as Standard model but contents, template and build directories are not all
contained in a single directory. This is useful if you need to build multiple
websites from the same content.

### Minimal project model
A shared content and template directory containing the build directory.

```
content and template files...
build/
    website files...
```
Minimal project model build command: `hindsite build -content . -template .`

This model may be useful for simple projects but it's not a good idea to mix
content and design like this if you plan to share or reuse your project
template.


## Content document markup
Content documents can be written in Markdown or Rimu.

### Rimu markup
The optional `config.rmu` file at the root of content directory will be
prepended to all Rimu files before they are rendered and mostly used to define
macro values.


## CLI
Usage (c.f. `go help`and `dlv help`): 

```
hindsite command [arguments]

The commands are:

    init
    build
    serve

Use "hindsite help [command]" for more information about a command.

hindsite init [-project PROJECT_DIR] [-template TEMPLATE_DIR] [-content CONTENT_DIR]
hindsite build [-clean] [-drafts] [-slugify] [-set VALUE] [-project PROJECT_DIR] [-template TEMPLATE_DIR] [-content CONTENT_DIR] [-build BUILD_DIR] -index [INDEX_DIR]
hindsite serve [-port PORT] [-project PROJECT_DIR] [-build BUILD_DIR]
hindsite -h | --help | help

```
<!-- - If content, template or build directory paths are relative they are relative
  to the project directory. -->
- `PROJECT_DIR` defaults to `.`
- `CONTENT_DIR` defaults to `content`
- `BUILD_DIR` defaults to `build`
- `TEMPLATE_DIR` defaults to `template`
- `INDEX_DIR` defaults to `BUILD_DIR/indexes`
- All commands support the `-v` (verbose) option).
- Build command supports the `-n` (dry run) option.

### Init command
Creates new project content directory from an existing template directory.

- Clones the `TEMPLATE_DIR` directory structure to `CONTENT_DIR`.
- Copies example content documents to `CONTENT_DIR`.
- Creates `CONTENT_DIR` directory (if the `CONTENT_DIR` already exists init will
  refuse to continue).


### Build command
Build static website from content and template directories.

- Include draft pages if the `-drafts` option is specified.
- Slugify content document paths if the  `-sligify` option is specified.
- The `-set` option `VALUE` is formated like `name=value`. It sets a named
  configuration parameter and template variable, for example
  `-set urlprefix=/blog`.
- If the `-clean` option is specified the the entire contents of the build
  directory is deleted, forcing all files to be rebuilt. By default only missing
  and modified files are processed.

Build sequence:

- Check for pathalogical content, template and build directories overlap.
- Check the template directory structure is mirrored in the content directory.
- Check that all directory paths and content document file names are slugified.
- Check configuration file validity.
- Delete contents of build directory (`-clean` build option).
- Copy static files from template to build directory.
  files).
- Copy static files from content to build directory (content directory files
  take precedence).
- Process all content documents.
- Process each indexed directory generating index and document web pages.

Corner cases examples:

```
# Template and content in ./docs output to ./build
hindsite build -content docs -template docs

# Template and content in current directory, output to ./build
hindsite build -content . -template .

```

### Serve command
Open a development server on the `BUILD_DIR` directory.

- The `-port PORT` option sets the server port (defaults to 1212).


## Implementation ideas
- Put binaries on BinTray?
- Add Disqus comments -- I suspect this is just a content templating issue.


### Implementation notes
- Internally all directory and file paths are absolute and platform specific (manipulated by the `path/filepath` package).

### Dates
- Template variables:
**shortDate, mediumDate, longDate**:
Publication date in the config file shortDateFormat, mediumDateFormat and longDateFormat formats.

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


## Testing
Need to test of example projects and then compare the contents of the
build directory to an `expected` directory for equality.

- This would provide a way to test arbitrarily complex projects.
- But approach is unwieldy for error scenarios.
- The test cases to include ./examples/
- Use these projects for benchmarks.
