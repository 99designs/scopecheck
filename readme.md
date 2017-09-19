scopecheck
==========

Checks for access to the possibly wrong var by comparing types. for example, this is nearly always wrong:

```go
var t1 testing.T

t1.Run("foo", func(t2 *testing.T) {
    t1.Fail()
})
```


It only looks at parameters from inline functions that have a similar type declared in its parent scope. There is a little bit of fuzzy matching so that an interface passed to the function can match a concrete type in the parent. eg 

```go
// is a *chi.Mux
r1 := chi.NewRouter()

r1.Group(func(r2 chi.Router) { // is an interface, chi.Router. Will still match
    r1.Use(nil)
})
``` 
