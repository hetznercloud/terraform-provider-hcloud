package testsupport

import "github.com/stretchr/testify/mock"

// GetIntChan tries to cast the result of args.Get(i) to a chan int.
//
// If the result of args.Get(i) is nil, nil is returned. If casting fails,
// GetIntChan panics
func GetIntChan(args mock.Arguments, i int) chan int {
	v := args.Get(i)
	if v == nil {
		return nil
	}
	return v.(chan int)
}

// GetErrChan tries to cast the result of args.Get(i) to a chan error.
//
// If the result of args.Get(i) is nil, nil is returned. If casting fails,
// GetErrChan panics
func GetErrChan(args mock.Arguments, i int) chan error {
	v := args.Get(i)
	if v == nil {
		return nil
	}
	return v.(chan error)
}
