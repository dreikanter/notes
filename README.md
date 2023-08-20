# Public notes automation

**Disclaimer:** This project is a single-purpose software built for personal use. It does not include a user manual and does not guarantee backward compatibility. I never intended to distribute this code, so there is no gemspec, public releases, changes history, or versioning. Very minimal customizability either.

## Usage

Setup:

```bash
brew install caddy
bundle
yarn && yarn build
```

Build everything:

```bash
NOTES_CONFIGURATION_PATH=~/src/notes/configuration.yml ~/src/notes/bin/build
```

Run local HTTP server to preview the site:

```bash
bin/serve
```
