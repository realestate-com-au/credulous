# this software is archived

As at 2021-03, REA Group no longer directly uses nor supports this software - as such, we have archived it.

You will note that the LICENSE file _already_ describes a lack of any warranty.

We leave the software visible as an example of our technology journey through the years, and hope it's useful for such.

(you might consider https://github.com/99designs/aws-vault)

-----

# Credulous

**credulous** is a command line tool that manages **AWS (IAM) Credentials
securely**. The aim is to encrypt the credentials using a user's **public
SSH Key** so that only the user who has the corresponding **private SSH
key** is able to see and use them. Furthermore the tool will also enable
the user to **easily rotate** their current credentials without breaking
the user's current workflow.

## Main Features

* Your IAM Credentials are securely encrypted on disk.
* Easy switching of Credentials between Accounts/Users.
* Painless Credential rotation.
* Enables rotation of Credentials by external application/service.
* No external runtime dependencies beyond minimal platform-specific
  shared libraries

## Installation

### For Linux (.RPM or .DEB packages)

Download your [Linux package](https://github.com/realestate-com-au/credulous/releases)


### For OSX

If you are using *[Homebrew](http://brew.sh/)* you can follow these steps to install Credulous

1. ```localhost$ brew install bash-completion```
1. Add the following lines to your ~/.bash_profile:
```
if [ -f $(brew --prefix)/etc/bash_completion ]; then
    . $(brew --prefix)/etc/bash_completion
fi
```
1. ```localhost$ brew install https://raw.githubusercontent.com/realestate-com-au/credulous-brew/master/credulous.rb```
1. Add the following lines to your ~/.bash_profile:
```
if [ -f $(brew --prefix)/etc/profile.d/credulous.sh ]; then
    . $(brew --prefix)/etc/profile.d/credulous.sh
fi
```

### Command completion

Command completion makes credulous much more convenient to use.

OSX: `brew install bash-completion`

Centos: [Enable EPEL repo and install bash-completion](http://unix.stackexchange.com/questions/21135/package-bash-completion-missing-from-yum-in-centos-6)

Debian/Ubuntu: bash-completion is installed and enabled by default. Enjoy!



## Usage

Credentials need to have the right to inspect the account alias, 
list access keys and examine the username of the user for whom they
exist. An IAM policy snippet like this will grant sufficient
permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PermitViewAliases",
            "Effect": "Allow",
            "Action": [ "iam:ListAccountAliases" ],
            "Resource": "*"
        },
        {
            "Sid": "PermitViewOwnDetails",
            "Effect": "Allow",
            "Action": [
                "iam:ListAccessKeys",
                "iam:GetUser"
            ],
            "Resource": "arn:aws:iam::*:user/${aws:username}"
        }
    ]
}
```

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
* [mercurial](http://mercurial.selenic.com)

Make sure you have [GOPATH](http://golang.org/doc/code.html#GOPATH) set in your environment

Download the dependencies

    $ go get -u # -u will update existing dependencies

Install [git2go](https://github.com/libgit2/git2go) (Optional if you already have it installed correctly in your environment)

    $ go get github.com/libgit2/git2go
    $ cd $GOPATH/src/github.com/libgit2/git2go && rm -rf vendor/libgit2
    $ git submodule update --init
    $ mkdir -p $GOPATH/src/github.com/libgit2/git2go/vendor/libgit2/install/lib
    $ make install
    # Run dependency update again for credulous
    $ cd $GOPATH/src/github.com/realestate-com-au/credulous && go get -u

Install the binary in your $GOBIN

    $ go install

## Tests

First we make sure we have our dependencies

    go get -t

Make sure goconvey is installed, else use

    go get -t github.com/smartystreets/goconvey

Just go into this directory and either

    goconvey
    < Go to localhost:8080 in your browser >

Or just run

    go test ./...

## Roadmap
See [here](https://github.com/realestate-com-au/credulous/wiki/Roadmap)

![Credulous Security](https://github.com/realestate-com-au/credulous/raw/master/site/credulous-security.png)
