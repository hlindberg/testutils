# testutils
Golang testutils contains useful utilities for writing golang tests

# get it

```
go get github.com/hlindberg/testutils
```

# use it

Simple case:

Using `CheckEqual` which treats numerical values as equal irrespective of type if they have the same
numerical value. (There are many other "Check" methods available).

```
import "github.com/hlindberg/testutils"

func TestSomething(t *testing.T) {
    testutils.CheckEqual(1, 1, t)
}
```

Case with iteration:

```
import "github.com/hlindberg/testutils"

func TestSomething(t *testing.T) {
    tester := testutils.NewTester(t)
    for i := 0; i <10; i++ {
        tester.At(i).CheckEqual(i, i)
}
```
