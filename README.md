# Hindsite website generator

hindsite is a static website generator. It builds static websites with optional
document and tag indexes from Markdown source documents (e.g. blogs posts,
newsletters, articles).

The goal is an intuitive, minimalist application that is easy to use and
understand. The number of features and concepts have been kept to a minimum.


## Quick Start
[Download hindsite](TODO) for your platform and create a fully functional blog and
newsletter website with just two hindsite commands:

1. Create a new Hindsite project directory and install the builtin blog
   template:

    mkdir myproj
    hindsite init myproj -builtin blog

2. Build the website:

    hindsite build myproj

To view the website in your browser:

1. Start the hindsite web server:

    hindsite serve myproj

2. Open your Web browser at http://localhost:1212

The best way to start learning hindsite is to browse the blog [project
directory](#projects).

Try editing content documents and templates and rebuilding. If you are familar
with Jekyll or Hugo you'll be right a home (hindsite can read Jekyll and Hugo
Markdown document front matter).