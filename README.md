Go katas are small [programs](https://github.com/orgs/gokatas/repositories) meant to show and teach Go programming techniques. They are written by experienced programmers - see the comments inside the katas. To understand and remember these techniques install the gokatas CLI tool

```
go install github.com/gokatas/gokatas@latest
```

and repeatedly do this simple practice cycle 🥋:

1. Choose a kata

```
gokatas -sortby lines -wide
git clone https://github.com/gokatas/<kata>.git
cd <kata>
```

2. Read the kata and try to understand it

```
code .
gokatas -explain <kata> # we ask ChatGPT here
```

3. Delete (some of) the kata and try to write it back

```
git diff
```

4. Track your progress to stay motivated

```
gokatas -done <kata>
```

It's important to practice regularly. Start by taking baby steps. After some time it will require much less will power to practice. And your programming moves will start looking simpler and smoother.

Sample ouput:

```
❯ gokatas -sortby last -report -wide
Name       Description                      Lines  Done  Last done     Standard library packages      URL
----       -----------                      -----  ----  ---------     -------------------------      ---
findgo     walking filesystems              51     6x    0 days ago    cmp filepath fs fstest         https://github.com/gokatas/findgo.git
fetch      HTTP client                      49     5x    3 days ago    fmt http io os time            https://github.com/gokatas/fetch.git
dup        duplicate lines in files         30     5x    3 days ago    fmt os strings                 https://github.com/gokatas/dup.git
direction  enumerated type with iota        45     4x    4 days ago    fmt rand                       https://github.com/gokatas/direction.git
clock      TCP time server                  38     6x    5 days ago    io log net time                https://github.com/gokatas/clock.git
lognb      non-blocking concurrent logging  103    7x    5 days ago    fmt os signal time             https://github.com/gokatas/lognb.git
err        errors are values                48     1x    9 days ago    fmt io os                      https://github.com/gokatas/err.git
boring     concurrency patterns             190    7x    9 days ago    fmt rand time                  https://github.com/gokatas/boring.git
books      sorted table in terminal         55     6x    9 days ago    fmt os sort strings tabwriter  https://github.com/gokatas/books.git
areader    io.Reader implementation         36     8x    11 days ago   bytes testing                  https://github.com/gokatas/areader.git
bcounter   io.Writer implementation         22     6x    11 days ago   fmt                            https://github.com/gokatas/bcounter.git
netcat     TCP client                       26     2x    83 days ago   io log net os                  https://github.com/gokatas/netcat.git
parselogs  loop over JSON logs              47     1x    89 days ago   errors fmt io json log         https://github.com/gokatas/parselogs.git
shift      simple cipher                    54     2x    95 days ago   bytes testing                  https://github.com/gokatas/shift.git
shop       HTTP server                      43     1x    96 days ago   fmt http log                   https://github.com/gokatas/shop.git
proxy      TCP middleman                    39     1x    96 days ago   io log net                     https://github.com/gokatas/proxy.git
lookup     loadbalanced STDIN processing    68     2x    96 days ago   bufio fmt net os strings sync  https://github.com/gokatas/lookup.git
google     building search engine           187    3x    138 days ago  fmt rand time                  https://github.com/gokatas/google.git

           Mar             Apr             May             Jun             Jul             Aug             Sep
       -   -   -   -   -   -   1   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -
Mon    -   -   -   -   -   4   2   1   2   -   2   -   1   1   -   1   -   -   -   -   -   -   -   -   -   1   1
       -   -   -   -   -   -   -   -   -   -   -   -   -   -   2   -   -   -   -   -   -   -   -   -   -   -   -
Wed    -   -   -   -   -   3   9   -   -   -   -   -   5   1   -   -   1   -   -   -   -   -   -   -   -   2   -
       -   -   -   -   3   2   -   2   5   -   1   1   1   -   1   -   -   -   -   -   -   -   -   -   2   1   -
Fri    -   -   -   -   2   -   2   -   -   -   1   3   -   -   1   -   -   -   -   -   -   -   -   -   -   2   -
       -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   -   3   -   -
```
