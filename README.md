# danicat.dev Personal Website

This repository contains the source code for my personal website and blog, available at danicat.dev. Here you will find articles and tutorials about software development, cloud computing, and other technology-related topics.

## Built With

*   [Hugo](https://gohugo.io/) - Static site generator
*   [Blowfish](https://blowfish.page/) - Hugo theme for building personal websites
*   Other Hugo modules as listed in `go.mod`.

## Getting Started

To run the website locally, you'll need to have Hugo installed. You can find installation instructions in the [official Hugo documentation](https://gohugo.io/getting-started/installing/).

Once Hugo is installed, follow these steps:

1.  Clone this repository:
    ```bash
    git clone https://github.com/danicat/danicat.dev.git
    ```
2.  Navigate to the project directory:
    ```bash
    cd danicat.dev
    ```
3.  Update Hugo modules:
    ```bash
    hugo mod tidy
    ```
4.  Start the Hugo development server:
    ```bash
    hugo server
    ```
    This will typically make the site available at `http://localhost:1313/`.

## Content Management

All website content, primarily blog posts, is located in the `content/posts` directory. Each post has its own directory, which contains an `index.md` for the English version and an `index.pt-br.md` for the Portuguese version.

To add a new post:

1.  Create a new directory for your post within `content/posts`.
2.  Inside the new directory, create an `index.md` file for the English content and an `index.pt-br.md` for the Portuguese content.
3.  You can use the Hugo CLI to create a new content file with the basic front matter:
    ```bash
    hugo new posts/my-new-post/index.md
    ```

## Customization

*   **Main Configuration**: The primary configuration for the site can be found in `config/_default/hugo.toml`. This includes settings for the title, theme, menus, and other site-wide parameters.
*   **Custom Styling**: Custom CSS rules are located in `assets/css/custom.css`. You can modify this file to change the site's appearance.

## Deployment

Building the site for deployment is done using the standard Hugo command:
```bash
hugo
```
This will generate the static site in the `public/` directory. The contents of this directory can then be deployed to any static web hosting service.

Many services like Netlify, Vercel, GitHub Pages, Cloudflare Pages, etc., can automatically build and deploy Hugo sites directly from a Git repository.

## Multi-language Support

The website is available in English and Portuguese. The language-specific configurations can be found in `config/_default/languages.en.toml` and `config/_default/languages.pt-br.toml`.

## License

This project is licensed under the terms of the MIT License. See the [LICENSE](LICENSE) file for details.
