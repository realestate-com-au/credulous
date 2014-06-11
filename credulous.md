% credulous(1) | Version ==VERSION==
% Colin Panisset, Mike Bailey et. al., REA Group
% Jun 2, 2014

# NAME

`credulous` - securely store and retrieve AWS credentials

# SYNOPSIS

`credulous <command> [<args>]`

# DESCRIPTION

Credulous manages AWS credentials for you, storing them securely
and retrieving them for placement in your shell runtime environment on
demand so that they can be used by other tools.

Credulous makes use of SSH RSA public keys to encrypt credentials at
rest, and the corresponding private keys to decrypt them. It supports
multiple AWS IAM users in multiple accounts, and provides the
capability to store custom environment variables encrypted along with
each set of credentials.

# COMMANDS

**save** Encrypt AWS credentials from the current environment
variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` with an SSH
RSA public key, and store them securely.

**source** Decrypt a set of AWS credentials for a given username and
account alias and make them available in a form suitable for eval'ing
into the current shell runtime environment.

**current** Query the AWS APIs using the current credentials and
display the username and account alias.

**rotate** Force a key rotation to occur -- credulous will delete one
key and create a new one, saving the new credentials into the repository.

**display** Show the currently loaded AWS credentials

**list** Show a list of all stored `username@alias` credentials.

# OPTIONS

**-h**
**--help**

> All commands take the `-h` or `--help` option to describe the command
> and its available full option set.

## Options for the save subcommand

**-k \<keyfile\>**
**--key \<keyfile\>**

> Specify the SSH public key to use in saving the current credentials.
> If more than one key is specified, the credentials will be saved
> multiple times, encrypted with each different public key.

**-e \<VAR\>=\<value\>**
**--env \<VAR\>=\<value\>**

> Save the environment variable `VAR` with the value `value` along with
> the encrypted credentials. The option can be used multiple times to
> save multiple different environment variables. All specified
> environment variables are encrypted alongside the credentials.

**-u \<username\>**
**--username \<username\>**

> Specify the AWS IAM username for the current credentials. Note that
> this is unnecessary if you have an active Internet connection, as
> credulous will query AWS for the username from the current
> credentials when `save` is called. If `username` is specified, you
> __must__ also specify **--account** and **--force** to prevent
> credulous querying AWS.

**-a \<account\>**
**--account \<account\>**

> Specify the AWS account alias. If you specify this option, you
> __must__ also specify **--username** and **--force**, otherwise
> credulous will query AWS for the account alias. If no account alias
> has been defined, credulous will use the numeric account ID instead.

**-f**
**--force**

> Do not attempt to verify the username or account alias/ID with AWS
> when saving. This is useful if you don't have an Internet connection
> when saving the credentials, but you __must__ specify both the
> username and account alias at the same time.

## Options for the source subcommand

If no options are specified, and no credential is specified on the
command-line, and only a single set of credentials have been saved,
credulous will read those credentials. If multiple credentials have
been saved, you will have to specify the credentials to source.

Note that if the SSH private key used to decrypt the credentials is not
protected with a passphrase, credulous will issue a warning.

**-k \<keyfile\>**
**--key \<keyfile\>**

> Use the specified SSH RSA private key to decrypt the credentials.

**-a \<account\>**
**--account \<account\>**

> Load credentials for the named account

**-u \<username\>**
**--username \<username\>**

> Load credentials for the named username.

**-f**
**--force**

> Do not attempt to verify that the loaded credentials match the
> username and account specified on the command-line. This is required
> if you have no Internet connection, BUT __represents a security
> risk__ in that a third party could substitute different credentials
> and you would have no way of knowing that this had happened until AWS
> API calls were made using those credentials, leading to a potential
> leakage of information, and financial or operational damage.

**-c \<username\>@\<account\>**
**--credentials \<username\>@\<account\>**

> Load the specified credentials. This is the default action for the
> `source` subcommand, so invocations like `credulous source foo@bar`
> are perfectly acceptable.

## Options for the current subcommand

There are no options for the `current` subcommand.

## Options for the rotate subcommand

**-k \<keyfile\>**
**--key \<keyfile\>**

> Specify the SSH public key to use in saving the new credentials.
> If more than one key is specified, the credentials will be saved
> multiple times, encrypted with each different public key.

**-e \<VAR\>=\<value\>**
**--env \<VAR\>=\<value\>**

> Save the environment variable `VAR` with the value `value` along with
> the encrypted credentials. The option can be used multiple times to
> save multiple different environment variables. All specified
> environment variables are encrypted alongside the credentials.

## Options for the display subcommand

There are no options for the `display` subcommand.

## Options for the list subcommand

There are no options for the `list` subcommand.

# EXAMPLES

## Save a set of AWS credentials from the current environment

    host$ env | grep AWS
    AWS_ACCESS_KEY_ID=AKIAJRETNBUEIZ3S6VU2
    AWS_SECRET_ACCESS_KEY=ffLbUThxWlKvR/Wp/qanXlgpthqipyDsUxHBUrN2
    host$ credulous save
    Saving credentials for hoopy@frood

## Save a set of AWS credentials using a specific SSH public key

    host$ credulous save -k /path/to/ssh/key.pub

## Save a set of environment variables along with the AWS credentials

    host$ credulous save -e AWS_DEFAULT_REGION=us-west-2 \
        -e FOO=bar -e BACON=yummy

## Load a particular set of credentials

    host$ credulous source hoopy@frood
    Enter passphrase for /path/to/my/ssh/privkey_rsa: ********
    export AWS_ACCESS_KEY_ID=AKIAJRETNBUEIZ3S6VU2
    export AWS_SECRET_ACCESS_KEY=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

## Place the sourced credentials into the runtime environment

    host$ eval $( credulous source hoopy@frood )
    Enter passphrase for /path/to/my/ssh/privkey_rsa: ********
    host$ env | grep AWS
    AWS_ACCESS_KEY_ID=AKIAJRETNBUEIZ3S6VU2
    AWS_SECRET_ACCESS_KEY=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

# AUTHORS

Colin Panisset, Mike Bailey, Greg Dziemidowicz, Paul van de Vreede,
Mujtaba Hussain, Stephen Moore.

# BUGS

Please report bugs via the GitHub page at
https://github.com/realestate-com-au/credulous/issues

# COPYRIGHT

Copyright (c) 2014 REA Group, Pty Ltd. `credulous` is distributed
under the MIT license: http://opensource.org/licenses/MIT
There is NO WARRANTY, to the extent permitted by law.
