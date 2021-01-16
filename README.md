# akyuu
Simple CDNish system for self hosted large uploads

## the gist
what needs to happen here is simple (bet):
* [*] implement file storage api (this can be migrated to GAE stuff later)
* [*] tie into database for managing entries
* [*] write rest-like api to interface
* [ ] """frontend"""
* [ ] get things to play nicely with discord/whatever
* [*] dockerfile
* [ ] scale

initial POC is just images -> other media -> link shortening -> text (as frontend progresses and all that)

obviously we need to be a bit smart when sending files, so the POST stuff below is a bit of an oversimplification. for POC it works though.
it'll probably need to work using multipart sending in the future

we can spec out the backend api to look like:
* POST      /i          -> imageId
* GET       /i/:imageId -> serve image file or 404
* DELETE    /i/:imageId -> delete image or 404

these all require a token of some sort, and also will exist for various things (/i/ = images, /v/ = videos, /g/ = gif, /t/ = text, etc)

text would end up having syntax highlighting and such

link shortening is also a plus:
* POST      /s          -> linkId
* GET       /s/:linkId  -> redirect to original link
* DELETE    /s/:linkId  -> unregister link. this should be a bit restricted to use

files/resources should be stored with some kind of UUID to prevent guessing of other items being stored

## running

```
docker build -t akyuu .
docker run --volume akyuu-vol:/akyuu/akyuu: -p 3000:80 akyuu
```
