fakes
=====
Files in this directory are utilized for testing the `health` library and can be
safely ignored by most.

If you are doing development work on the lib and are introducing a change to some
interface, you will need to regenerate one or more fakes. To do so, make sure you
have [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) installed
and run `go generate`.

Note that you _will_ have to modify the generated files and remove the import of
the `health` library itself (as that will cause circular import problems).