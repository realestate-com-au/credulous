# Things needed to build on OSX

## Mavericks

* developer command-line tools
    * can get just by trying to run `git` and then following the
      prompts to download command-line tools -- do not need full XCode
* go
    * download from
      http://golang.org/dl/go1.2.2.darwin-amd64-osx10.8.pkg and run the
      installer
* hg
    * download from http://mercurial.selenic.com/mac/binaries/Mercurial-3.0.1-py2.7-macosx10.9.zip
    * it's an unsigned package, so will have to go to System Preferences
      -> Security & Privacy and allow "Anywhere" for downloaded apps
* pkg-config
    * yeah, you need [Homebrew](http://brew.sh)
    * `brew install pkg-config`
* libgit2
    * `brew install https://raw.githubusercontent.com/realestate-com-au/credulous-brew/master/libgit2-0.21.0.rb`
   
