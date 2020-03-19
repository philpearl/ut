package testcode

/*
	These interface definitions are here for the `TestNestedInterfaces`
	To check the generating nested interfaces works
*/

type Interface1 interface {
	Method1(value1 string) error
}

type Interface2 interface {
	Method2(value2 string) error
}

type Interface3 interface {
	Interface1
	Interface2
	Method3(value3 string) error
}

type Interface4 interface {
	Interface3
	Method4(value4 string) error
}

type Interface4Impl struct {
}

func (i *Interface4Impl) Method1(value1 string) error {
	return nil
}

func (i *Interface4Impl) Method2(value2 string) error {
	return nil
}

func (i *Interface4Impl) Method3(value3 string) error {
	return nil
}

func (i *Interface4Impl) Method4(value4 string) error {
	return nil
}
