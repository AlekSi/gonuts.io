Filling issues and contributing to gonuts.io
============================================

First of all, thank you for your interest in making packaging better! There are number of ways you can help:

* reporting bugs;
* proposing features;
* contributing code bug fixes and new features;
* contributing documentation fixes (there is probably a ton of grammar errors :/) and improvements.

The following sections describes those scenarios. Golden rule: communicate first, code later.

Reporting bugs
--------------

1. Enable debug output: set "Debug" to "true" in `~/.nut.json`.
2. Make sure bug is reproducible with latest released version: `go get -u github.com/AlekSi/nut/...`.
3. Search for [existing bug report](https://github.com/AlekSi/nut/issues).
4. Create a new issue if needed. Please do not assign any label.
5. Include output of:

		(cd $GOPATH/src/github.com/AlekSi/nut && git describe --tags)
		go env

6. Include any other information you think may help.

Proposing features
------------------

Please add your comments to [existing feature requests](https://github.com/AlekSi/gonuts.io/issues?labels=feature), but do not create new without proposing them in [mailing list](https://groups.google.com/group/gonuts-io) first.

Contributing changes
--------------------

1. Read all previous sections first.
2. You can make small changes right in the web interface. Spot a typo? Fix it! :)
3. For bigger changes make a fork on GitHub and clone it into `$GOPATH/src/github.com/AlekSi/gonuts.io`.
4. Run `make prepare` to install remote packages and `make run` to run local server.
5. Create a topic branch.
6. Make your changes.
7. Publish your topic branch.
8. Make a pull request.
