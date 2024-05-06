Go katas is a [set](https://github.com/orgs/gokatas/repositories) of small Go programs written by experts. They contain techniques that you can re-use when writing Go code. But first you must learn them. And then remember them. The practice cycle to internalize the techniques is straightforward:

0. Install the gokatas tool

```
go install github.com/gokatas/gokatas@latest
```

1. Select one of the katas

```
gokatas -sortby lines -wide
git clone https://github.com/gokatas/bcounter.git # or git@github.com:gokatas/bcounter.git
cd bcounter
```

2. Read the documentation and code and try to understand it

```
code .
```

3. Delete the code in a file and try to write it back

```
git diff
```

4. To keep yourself motivated track what you have done

```
gokatas -done bcounter
```

It's important to practice regularly because repetition creates habits, and habits are what enable mastery. Start by taking baby steps. Set a goal that you are able to meet and insert it into your daily routines. After some time it will require much less will power to practice. Your programming moves will start looking simpler and smoother. 🥋
