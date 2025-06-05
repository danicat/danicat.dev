# danicat.dev Personal Website

This is the source code for the personal website and blog available at [danicat.dev](https://danicat.dev/). It is built using the [Hugo static site generator](https://gohugo.io/).

## Built With

*   [Hugo](https://gohugo.io/) - Static site generator
*   [Hugo Bootstrap Module](https://github.com/hugomods/bootstrap)
*   [Hugo Images Module](https://github.com/hugomods/images)

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

All website content, primarily blog posts, is located in the `content/posts` directory.

To add a new post:

1.  You can use the Hugo CLI to create a new content file:
    ```bash
    hugo new posts/my-new-post.md
    ```
2.  This will create a new Markdown file in `content/posts/my-new-post.md` with the basic front matter.
3.  Edit the new Markdown file to add your content.

## Customization

*   **Main Configuration**: The primary configuration for the site can be found in `hugo.yaml`. This includes settings for the title, theme, menus, and other site-wide parameters.
*   **Custom Styling**: Custom CSS rules are located in `static/css/custom.css`. You can modify this file to change the site's appearance.

## Deployment

Building the site for deployment is done using the standard Hugo command:
```bash
hugo
```
This will generate the static site in the `public/` directory. The contents of this directory can then be deployed to any static web hosting service.

Many services like Netlify, Vercel, GitHub Pages, Cloudflare Pages, etc., can automatically build and deploy Hugo sites directly from a Git repository.

## License

This project is licensed under the terms of the MIT License. See the [LICENSE](LICENSE) file for details.
