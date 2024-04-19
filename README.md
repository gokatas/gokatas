> I fear not the man who has practiced 10,000 kicks once, but I fear the man who has practiced one kick 10,000 times. -- Bruce Lee

Katas (形) are practiced in martial arts as a way to internalize and perfect techniques, so they can be executed and adapted under various circumstances without much hesitation. Let's try something similar with Go programming. Go katas is a small set of [repositories](https://github.com/orgs/gokatas/repositories) containing small Go programs that are practical and relatively simple. The practice workflow is straightforward:

0. Install the gokatas tool

```
go install github.com/gokatas/gokatas@latest
gokatas -h
```

1. Select one of the katas and clone it (you might want to start with the smallest one)

```
gokatas -sortby lines
mkdir -p ~/github.com/gokatas && cd ~/github.com/gokatas
git clone https://github.com/gokatas/bcounter.git # or git@github.com:gokatas/bcounter.git
cd bcounter
```

2. **Read** the documentation and code and try to **understand** it (use included links, search engine or AI if you get stuck)

```
go doc -all
go doc io.Writer
vi main.go # or code .
```

3. **Delete** the code in a file and try to **write** it back. Check how you are doing

```
go run main.go
go test # in folders where <package>_test.go is present
git diff
```

Repeat the read-understand-delete-write cycle until you can write the kata from scratch.

It's important to practice regularly because repetition creates habits, and habits are what enable mastery. Start by taking baby steps. Set a goal that you are able to meet and insert it into your daily routines. After some time it will require much less will power to practice. Your programming moves will start looking simpler and smoother. 🥋
