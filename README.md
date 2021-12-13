# go-woodpecker (WIP)

a collection of tools to help developers deal with vulnerabilities

## prerequisites

* Maven v3.0.0+

## commands

```bash
woodpecker -h
woodpecker tree # shows depedency tree with vulnerabilities
```

## Maven projects (TODO)

Basically it simply does the following: 
(note that it use [go-git](https://github.com/go-git/go-git) instead of the usual git client)

* `mvn versions:use-next-releases`
* `mvn verify`
* `git branch -b auto-update-deps`
* `git add **pom.xml`
* `git commit -m "auto update dependencies"`
* `git push --set-upstream=auto-update-deps`
* create pull request


## Caveats
[ ] multi-modules project
[ ] dependency suite (dependencies share the same version)
