# Contributions #

Contributions to the Autograder project is always welcome. In order to get the
best out of contributions from everyone please follow the guidelines given
below.

Any contribution is welcome. If you see a thing which the project can benefit
from, open a issue or pull request. We appreciate small changes or optimizations
to code or language just as much as larger ones.

Normal procedure for code contributions:
1. Fork the main repository
2. If you are working on one of the issues, mark the issue with an assignment or
  leave a comment about it.
3. Implement your changes/additions
4. Open a pull request to the main repository

If the contribution is of good quality we pull it into the main repository.


NB: if you are editing any of the static files in the web package, remember to
run `go generate` before sending your pull request. This will build in all the
static files into the go code.

Command:
```
  go generate ./...
```

NB: Students which have a semester project with Autograder will need to follow
these guidelines as well. If the work done in this project holds a good enough
quality will be pulled into the main repository through a pull request.

The import paths when forking can be a bit tricky to handle. More advanced ways
to fix this are possible, but follow these steps for a easy solution to the
import problem:
1. Fork the main repository to you github account.
2. Use `go get` on your fork. `go get github.com/yourusername/autograder`
3. Change the folder named after your github username to `hfurubotten` in the go
source path structure.  
4. Implement your changes/addidtions.
5. Upload as normal. The link to your fork on github will still be in your git
configuration.
6. Open a pull request with changes.
