Go katas are small [programs](https://github.com/orgs/gokatas/repositories) written by experienced programmers. They contain techniques that you can reuse for Go coding. But first you must learn them. And then remember them. The practice cycle to internalize the techniques is straightforward:

0. Install the gokatas tool

```
go install github.com/gokatas/gokatas@latest
```

1. Choose a kata

```
gokatas -sortby lines -wide

git clone https://github.com/gokatas/<kata>.git
cd <kata>
```

2. Read the documentation and code and try to understand it

```
code .
```

3. Delete (some of) the code and try to write it back

```
git diff
```

4. Track your progress to stay motivated

```
gokatas -done <kata>
```

It's important to practice regularly because repetition creates habits, and habits are what enable mastery. Set a goal that you are able to meet and insert it into your daily routines. Start by taking baby steps. After some time it will require much less will power to practice. Your programming moves will start looking simpler and smoother. 🥋
