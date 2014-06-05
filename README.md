# Credulous

**credulous** is a command line tool that manages **AWS (IAM) Credentials securely**. The aim is
to encrypt the credentials using a user's **public SSH Key** so that only the user who has the
corresponding **private SSH key** is able to see and use them. Furthermore the tool will also
enable the user to **easily rotate** their current credentials without breaking the user's current
workflow.

## Main Features

* Your IAM Credentials are securely encrypted on disk.
* Easy switching of Credentials between Accounts/Users.
* Painless Credential rotation.
* Enables rotation of Credentials by external application/service.
* No external runtime dependencies beyond minimal platform-specific shared libraries

## Installation

### For Linux (.RPM or .DEB packages)

Download your [Linux package](https://github.com/realestate-com-au/credulous/releases)


### For OSX

If you are using homebrew you can follow these steps to install Credulous
1. ```$ brew install bash_completion```
1. ```$ brew install https://raw.githubusercontent.com/realestate-com-au/credulous-brew/master/credulous.rb```
1. Add the following lines to your ~/.bash_profile:
```
if [ -f $(brew --prefix)/etc/profile.d/credulous.sh ]; then
    . $(brew --prefix)/etc/profile.d/credulous.sh
fi
```

#### Bash Completion (OSX)

Bash completion for subcommands and credentials is available.

OSX users will need to `brew install bash_completion`


## Usage

You can have a [look at the manual
page](https://github.com/realestate-com-au/credulous/blob/master/credulous.md), if that's your thing.

Storing your current credentials in Credulous

    $ export AWS_ACCESS_KEY_ID=YOUR_AWS_ID
    $ export AWS_SECRET_ACCESS_KEY=XXXXXXXXXXX
    $ credulous save # Will ask credulous to store these credentials
    # saving credentials for user@account

Displaying a set of credentials from Credulous

    $ credulous source -a account -u user
    export AWS_ACCESS_KEY_ID=YOUR_AWS_ID
    export AWS_SECRET_ACCESS_KEY=XXXXXXXXXXX


## Development

[![Build Status](https://travis-ci.org/realestate-com-au/credulous.svg)](https://travis-ci.org/realestate-com-au/credulous)

Required tools:
* [go](http://golang.org)
* [git](http://git-scm.com)
* [bzr](http://bazaar.canonical.com)

Make sure you have [GOPATH](http://golang.org/doc/code.html#GOPATH) set in your environment

Download the dependencies

    $ go get -u # -u will update existing dependencies

Install the binary in your $GOBIN

    $ go install

## Tests

First we make sure we have our dependencies

    go get -t

Just go into this directory and either

    goconvey
    < Go to localhost:8080 in your browser >

Or just run

    go test ./...

## Roadmap
See [here](https://github.com/realestate-com-au/credulous/wiki/Roadmap)

![Credulous Security](https://github.com/realestate-com-au/credulous/raw/master/site/credulous-security.png)
