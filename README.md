# checksubdomains

Finds HTTP-reachabe subdomains of a given host using [sublist3r](https://github.com/aboul3la/Sublist3r).

If you provide a file with the `-out_html` argument or simply test `--html`, you'll get an HTML file to help navigate the subdomains, e.g. this [example](https://spudtrooper.github.io/checksubdomains/nasa.gov.html). You can traverse with the left/right arrow keys.

## Example usage

Install:

```
go get -u github.com/spudtrooper/checksubdomains
```

Check the subcomains of `foo.com` where the sublist3r main file is `/path/to/sublist3r.py`:

```
checksubdomains --host foo.com --sublist3r /path/to/sublist3r.py --html
```