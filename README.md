TODO:
- prepare proper makefiles
- control root user on containers
- return pointers for New functions
- use pointers when implementing functions for structs
- upgrade to go 1.26 and use errors.AsType 
// using errors.AsType
if target, ok := errors.AsType[AppError](err); ok {
    fmt.Println("application error:", target)
}