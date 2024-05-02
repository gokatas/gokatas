Go katas is a [set](https://github.com/orgs/gokatas/repositories) of small Go programs written by experts. They contain techniques that you can re-use when writing Go code. The practice workflow to internalize the techniques is straightforward:

0. Install the gokatas tool

```
go install github.com/gokatas/gokatas@latest
```

1. Select one of the katas and clone it (you might want to start with the smallest one)

```
gokatas -sortby lines -wide
git clone https://github.com/gokatas/bcounter.git # or git@github.com:gokatas/bcounter.git
cd bcounter
```

2. **Read** the documentation and code and try to **understand** it (use included links, search engines or AI if you get stuck)

```
code .
```

3. **Delete** the code in a file and try to **write** it back. Check how you are doing

```
go run main.go
go test # in folders where <package>_test.go is present
git diff
```

4. To keep yourself motivated track what you have done

```
gokatas -done bcounter
```

It's important to practice regularly because repetition creates habits, and habits are what enable mastery. Start by taking baby steps. Set a goal that you are able to meet and insert it into your daily routines. After some time it will require much less will power to practice. Your programming moves will start looking simpler and smoother. 🥋
