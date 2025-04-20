# Tinylink

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/username/project/actions)

URL Shortener

<!-- ## Table of Contents

- [Tinylink](#tinylink)
	- [Table of Contents](#table-of-contents)
	- [Goals](#goals)
	- [Features](#features)
---

## Goals
- DDD pattern
- Implement stateless auth - refresh (long-lived) / access (short-lived JWT) tokens. Use redis to store refresh tokens
- SQLite for persisted, redis for tinylinks
- Analytics - track clicks, geolocation, etc...
---

## Features
- Password protected links [NO]
- Browser extension - single click to get tinylink
- Tinylinks - short URLs
	- they are persisted if user creates them
	- they expire after 30 days if anonymous (non-auth user) creates them
	- ??? when they expire, should i make hashes available again ???
	- ??? they become persisted when user creates an account - need to be added via ???
	- redis cached for 6 hours when accesed (avoids db call for frequently accessed tinylinks)
	- Support for bulk inserts (multipart-form for json/yaml/xml/csv) - only for authenticated users