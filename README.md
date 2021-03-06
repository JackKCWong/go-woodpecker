# go-woodpecker (WIP)

Give developers the last-mile help in fixing vulnerabilities  

## prerequisites

* Maven v3.1.0+ (required by [Maven Dependency-Check Plugin](https://jeremylong.github.io/DependencyCheck/dependency-check-maven/index.html))

## commands

```bash
woodpecker -h
woodpecker tree # shows depedency tree with vulnerabilities
woodpecker kill cve_id # updates the dependency until the cve_id is fixed. does NOT work with multi-module projects
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


# License

[MIT](https://github.com/JackKCWong/go-woodpecker/blob/main/LICENSE)
