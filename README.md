Katas (形) are practiced in martial arts as a way to internalize and perfect techniques, so they can be executed and adapted under various circumstances without much hesitation. Let's try something similar with Go programming.

Go katas is a small set of [repositories](https://github.com/orgs/gokatas/repositories) containing Go programs that are correct (true), practical (good), clear and simple (beautiful). The practice workflow is straightforward:

0. Install the gokatas tool
```
go install github.com/gokatas/gokatas@latest
```
1. Select one of the Go katas and clone it, for example
```
gokatas
git clone git@github.com:gokatas/bcounter.git
cd bcounter
```
2. **Read** the documentation and code and try to **understand** it (use included links, search engines or AI)
```
go doc [io.Writer]
vim main.go
```
3. Delete the code in a file and try to **write** it back. Check how you are doing
```
go test    # for packages with tests (package_test.go)
go run     # for main packages (main.go)
git diff
```

Repeat steps 2 and 3 until you feel comfortable with the kata. At that point stop practicing it and go back to step 1.

> Serva ordinem et ordo servabit te.

It's important to practice regularly because repetition creates habits, and habits are what enable mastery. Start by taking baby steps. Set a goal that you can meet, for example 10 minutes every day before lunch. At first it's fine even if you only read through one of the katas. After some time it will require much less will power to practice. Your programming moves will start looking simpler and smoother. 🥋
