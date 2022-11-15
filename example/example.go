package example

//go:generate genmock -package=github.com/philpearl/ut/example -interface=Fred -mock-package=example

type George struct{}

type Fred interface {
	sanit(blah string)
	iit(fred any)
	many(things ...string)
	doit(blah string) int
	donit(blah, fah string) (int, error)
	adonit(blah, fah George, brian func(int) error) (an int, err error)
}

func DoSomething(f Fred) {
	f.sanit("cheese")
	if f.doit("lemons") > 4 {
		f.many("a", "b")
	}
}
