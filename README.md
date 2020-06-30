# go.expiremap

Package expiremap is a simple thread-safe cache like sync.Map. Value in this map will automatically be deleted after expiration time.

This package acts like go-cache(github.com/patrickmn/go-cache), but much faster. The limitations are: expiration could not changed once the map is established, and the actual expiration time for each value may have an error of up to one second. These are the cost of increased effenciency.
