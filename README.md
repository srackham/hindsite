# Hindsite website generator

Hindsite is a static website generator. It's easy to learn an use and very fast.
It builds static websites with optional document and tag indexes (e.g. blogs
posts, newsletters, articles) from Markdown source documents.


## Quick Start
1. Download the Hindsite executable for your platform.

2. Create a new Hindsite project directory and install the builtin blog
   template:

        mkdir myproj
        hindsite init myproj -builtin blog

3. Build the website:

        hindsite build myproj

4. Start the hindsite web server:

        hindsite serve myproj

5. Open your Web browser at http://localhost:1212

To customise the website edit the project's Markdown documents (`.md` files in
the project `content` directory), templates (`.html` files in the project
`template` directory) and configuration files (`config.{toml,yaml}` in the
project `template` directory).
  
If you are familar with Jekyll or Hugo you'll be right a home (hindsite can read
Jekyll and Hugo Markdown documents).

