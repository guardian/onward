# Guardian Onward

**PROTOTYPE / WIP**

HTTP JSON API to serve required metadata for Guardian onward components (lists
of content such as Most Viewed, or Story Packages).

Edition endpoints are cached for 10 minutes in memory but other data is not; the
assumption is that most caching happens at the edge (CDN) level.
