legit
-----

A git web frontend written in Go.

Pronounced however you like; I prefer channeling my inner beret-wearing
Frenchman, and saying "Oui, il est le git!"

But yeah it's pretty legit, no cap on god fr fr.


FEATURES

• Fully customizable templates and stylesheets.
• Cloning over http(s).
• Less archaic HTML.
• Not CGI.


INSTALLING

Clone it, 'go build' it.


CONFIG

Uses yaml for configuration. Looks for a 'config.yaml' in the current
directory by default; pass the '--config' flag to point it elsewhere.

Example config.yaml:

    repo:
      scanPath: /var/www/git
      readme:
        - readme
        - README
        - readme.md
        - README.md
      mainBranch:
        - master
        - main
      ignore:
        - foo
        - bar
    dirs:
      templates: ./templates
      static: ./static
    meta:
      title: git good
      description: i think it's a skill issue
    server:
      name: git.icyphox.sh
      host: 127.0.0.1
      port: 5555

These options are fairly self-explanatory, but of note are:

• repo.scanPath: where all your git repos live (or die). legit doesn't
  traverse subdirs yet.
• dirs: use this to override the default templates and static assets.
• repo.readme: readme files to look for.
• repo.mainBranch: main branch names to look for.
• repo.ignore: repos to ignore, relative to scanPath.
• server.name: used for go-import meta tags and clone URLs.


NOTES

• Run legit behind a TLS terminating proxy like relayd(8) or nginx.
• Cloning only works in bare repos -- this is a limitation inherent to git. You
  can still view bare repos just fine in legit.
• The default head.html template uses my CDN to fetch fonts -- you may
  or may not want this.
• Pushing over https, while supported, is disabled because auth is a
  pain. Use ssh.
• Paths are unveil(2)'d on OpenBSD.


IDEAS

• "Private" repos only available over Tailscale.
• Support or cgit-like filters (for readmes etc.).


LICENSE

legit is licensed under MIT.
