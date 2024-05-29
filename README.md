# Public notes automation

**Disclaimer:** This project is _single-purpose software_ built for personal use. It is not a product, does not include a user manual, and does not guarantee backward compatibility. I don't intend to distribute this code, so there is no gemspec, public releases, change log, or versioning. There is very minimal customizability by design.

## Usage

Setup:

```bash
brew install caddy
bundle
yarn
```

Upgrade packages:

```bash
yarn upgrade --latest
```

Run the tests:

```bash
rspec
```

Build CSS:

```bash
yarn build
```

Build the site:

```bash
NOTES_CONFIGURATION_PATH=~/src/notes/configuration.yml ~/src/notes/bin/build
```

Run local HTTP server to preview the site:

```bash
bin/serve
```
