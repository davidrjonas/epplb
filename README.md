epplb
=====

[![Software License][ico-license]](LICENSE.md)

A load-balancing, connection-caching, reverse proxy for [EPP](https://tools.ietf.org/html/rfc5730). Scratches the itch of maintaining connections to the registry with keepalives while clients connect and disconnect per command locally.

Authentication is not required per connection. This means that once a successful login has happened, new login will appear to succeed regardless of their clID or pw and commands will not be truly authenticated. Protect this proxy well or implement client auth comparison.

Usage
-----

This project is not yet complete.

TODO
----

- [ ] Add config file
- [ ] Multi proxies
- [ ] Stop keepalive ticker without logout, on connection problem
- [ ] Add expvar stats
- [ ] [Error wrapping](https://github.com/pkg/errors)
- [ ] Research possible partial read/writes in ReadFrame, WriteFrame
- [ ] Client auth comparison, client auth scheme

Future Improvements
-------------------
- Add an HTTP endpoint so clients can speak HTTP instead of [rfc5734](https://tools.ietf.org/html/rfc5734)
- Pipelining


License
-------

The MIT License (MIT). Please see [License File](LICENSE.md) for more information.

[ico-license]: https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square
