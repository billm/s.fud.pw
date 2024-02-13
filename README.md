NOTE - there is much security missing from this code, you probably don't want to use it :) 


# s.fud.pw

A URL shortener hosted at [http://s.fud.pw/](http://s.fud.pw/) built around requirements for offensive operations.

Features:
* [X] Clean / Dirty URL toggling 
  * [X] allow a "good" url to be tested with a redirect to a "malicious" url for subsequent requests
  * [X] Limit number of dirty url responses
* [ ] URL expiration



# TODO
* [ ] Make slugs table unique on slug
* [ ] Need slug hit counter
* [X] Need multiple URLs for a slug
* [ ] TOCTOU checkbox maybe?
* [ ] Created timestamp for slug
* [ ] Select recent slugs from table to show
* [X] Remove Heroku boilerplate from readme
* [ ] Better error handling on SQL failures (UI integration, shouldn't be info alerts!)
* [ ] http://s.fud.pw/r/pq: duplicate key value violates unique constraint "slugs_slug_key"
