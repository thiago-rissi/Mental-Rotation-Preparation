Web application
===============

[index.html](index.html) is the start page of the browser version of the experiment, published at <https://thiago-rissi.github.io/Mental-Rotation-Preparation/experiment/>. It asks for the participant number and starts the program.

The program itself is the one in [experiment](../experiment), compiled to WebAssembly. `experiment/build.sh web` puts this page, the compiled program and the SDL library together in one folder, and [pages.yml](../.github/workflows/pages.yml) does the same before publishing the site.
