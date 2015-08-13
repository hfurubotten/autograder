Autograder
==========
[![GoDoc](https://godoc.org/github.com/hfurubotten/autograder?status.svg)](http://godoc.org/github.com/hfurubotten/autograder)
[![Build Status](https://travis-ci.org/hfurubotten/autograder.svg?branch=master)](https://travis-ci.org/hfurubotten/autograder)
[![Coverage Status](https://coveralls.io/repos/hfurubotten/autograder/badge.svg?branch=master&service=github)](https://coveralls.io/github/hfurubotten/autograder?branch=master)

Autograder is an automatic feedback system for the students. It is integrated
with GitHub and manages courses and students within GitHubs git management
system. When students push code to their repositories, the push triggers a
continuous integration process. Results form this integration process is then
given to the students on their personal web pages. Teachers can then access this
integration log, thus have a valuable tool in the grading of lab assignments.

## Features ##
Listed below is some of the features in autograder.

### Training in industrial grade tools ###
The teaching environment of autograder is infact GitHub itself, thus training the
students in using version control systems to have control over their code and
assignments. Integrated in autograder is also a custom made continuous
integration tools. Version controlling and continuous integration is tools widely
used by the industry. Training the students in tools like git, GitHub and CI
makes the students more equipped when making the transition to working life.

### Automatic assignment testing ###
Students submit their assignments to GitHub and they are done with the hand in
procedure. After the code from the students have been uploaded, Autograder takes
over and starts the testing and the correction process of assignments automatically.

### Code reviewing ###
Within the assignment process the students or groups are have separated
repositories to work in. This is to let the students come up with different
solutions and not cheat from one another. However to let the students help and
learn from each other a code review system is built in. The students have access
to a code review repository and can with the application upload snippets of code,
and let the other students look over it.

### Awardsystem for online discussions ###
There is a great potential for students learning from each other and helping
each other to understand course material better. To award students efforts to
collaborate when learning a game engine has been built in. For actions done by
the students in written form on GitHub, either it is commenting on code or on a
issue, the students earn points. These points go to their profile and also count
on the specific course. The students can also rise in level and earn different
badges. These points can then be used by the teaching staff to assess how active
the student are, in helping other students and improving the course.

## Installation and configuration ##
How to install and configure Autograder from source is explained in our
[install][] file

### Supported Operating systems ###

Autograder has been tested on and support following operating systems:

- Ubuntu

## Contributions ##
We encourage contributions in order to have the best system possible. Before
starting on your contribution efforts, please read your
[contributions][contribute] file.

## Dependencies ##

When compiling these following libraies need to be included;
- [goauth2][]
- [go-github][]
- [go-dockerclient][]
- [boltdb][]

The runtime of the test enviorment are virtualized in [docker].

[goauth2]: http://golang.org/x/oauth2
[go-github]: https://github.com/google/go-github
[docker]: https://www.docker.com/
[go-dockerclient]: https://github.com/fsouza/go-dockerclient
[boltdb]: https://github.com/boltdb/bolt
[diskv]: https://github.com/hfurubotten/diskv
[contribute]: https://github.com/hfurubotten/autograder/blob/master/CONTRIBUTIONS.md
[install]: https://github.com/hfurubotten/autograder/blob/master/INSTALL.md
