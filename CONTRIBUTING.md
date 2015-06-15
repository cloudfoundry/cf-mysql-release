# Contributing to cf-mysql-release

## Contributor License Agreement

Follow these steps to make a contribution to any of CF open source repositories:

1. Ensure that you have completed our CLA Agreement for
   [individuals](http://www.cloudfoundry.org/individualcontribution.pdf) or
   [corporations](http://www.cloudfoundry.org/corpcontribution.pdf).

1. Set your name and email (these should match the information on your submitted CLA)

        git config --global user.name "Firstname Lastname"
        git config --global user.email "your_email@example.com"

1. Submit a Pull Request
    1. Fork the repository on github

    1. Update submodules (`./update`)

    1. Create a feature branch (`git checkout -b awesome_sauce`)
        * Run the unit tests to ensure that your local environment is working `./scripts/test-unit`

    1. Make changes on the branch:
        * Adding a feature
          1. Add tests for the new feature
          1. Make the tests pass
        * Fixing a bug
          1. Add a test/tests which exercises the bug
          1. Fix the bug, making the tests pass
        * Refactoring existing functionality
          1. Change the implementation
          1. Ensure that tests still pass
            * If you find yourself changing tests after a refactor, consider refactoring the tests first

    1. Run the [acceptance tests](docs/acceptance-tests.md) (update them if required).

    1. Commit your changes (`git commit`)
        * Small changes per commit with clear commit messages are preferred.

    1. Push to your fork (`git push origin awesome_sauce`)

    1. Submit a pull request in github, selecting `develop` as the target branch.
