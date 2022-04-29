# checksubdomains

Finds HTTP-reachabe subdomains of a given host using [sublist3r](https://github.com/aboul3la/Sublist3r).

## Example usage

Install:

```
go install github.com/spudtrooper/checksubdomains
```

Check the subcomains of `foo.com` where the sublist3r main file is `/path/to/sublist3r.py`:

```
checksubdomains --host foo.com --sublist3r /path/to/sublist3r.py
```