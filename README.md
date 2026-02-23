TODO:
- prepare proper makefiles
- use pointers when implementing functions for structs
- write documentation
- upgrade to go 1.26 and use errors.AsType 
// using errors.AsType
if target, ok := errors.AsType[AppError](err); ok {
    fmt.Println("application error:", target)
}
