Go App
======

[![Build Status](https://travis-ci.org/rande/goapp.svg?branch=master)](https://travis-ci.org/rande/goapp)

 - Try to normalize how an application should start without providing any convention about how each steps should be used.
 - Provide an application container to store services


Features
--------

 - Application container
 - Load configuration file (as string) and replace {{ env 'ENV' }} with env variables
 - Application lifecycle management: from load to exit
