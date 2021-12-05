# go-woodpecker (WIP)

A naive idea to keep advancing project dependencies versions until the number of vulnerabilities drops. 

## prerequisites

* Maven v3.8.5

## Maven projects

Basically it simply does the following: (note that it use go-git instead of the usual git client)

* `mvn versions:use-next-versions`
* `mvn verify`
* `git branch -b auto-update-deps`
* `git add **pom.xml`
* `git commit -m "auto update dependencies"`
* `git push --set-upstream=auto-update-deps`
* create pull request

