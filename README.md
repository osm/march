# march

March is a media archive service, you feed it with a URL and it downloads and
stores the content on your server.

## Example

```sh
# Archive the file.
$ curl -ufoo:foo -durl=https://www.openbsd.org/images/openbsd-logo.gif http://localhost:8080/foo
4bfc3395-0761-45d9-9e0b-9bccfa8df95d

# Get the file.
$ curl -s -oopenbsd-logo.gif http://localhost:8080/foo/4bfc3395-0761-45d9-9e0b-9bccfa8df95d

# Verify that the downloaded file has the same md5sum as the one in the
# archive.
$ md5sum openbsd-logo.gif
44cbb1c1332c1bc25070f89363b7241f  openbsd-logo.gif
44cbb1c1332c1bc25070f89363b7241f  archives/foo/4bfc3395-0761-45d9-9e0b-9bccfa8df95d
```
