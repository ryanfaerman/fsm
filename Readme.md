Why bother going through the trouble of running the guards in parallel? Consider the benchmarks below:

```shell
$ go test -bench=.
PASS
BenchmarkRulesetParallelGuarding    100000       13261 ns/op
ok  github.com/ryanfaerman/fsm 1.525s

$ go test -bench=.
PASS
BenchmarkRulesetSerialGuarding         1  1003140956 ns/op
ok  github.com/ryanfaerman/fsm 1.007s
```


```shell
$ go test -bench=.
PASS
BenchmarkRulesetParallelGuarding      100000   13163 ns/op
BenchmarkRulesetTransitionPermitted  1000000    1805 ns/op
BenchmarkRulesetTransitionDenied    10000000     284 ns/op
BenchmarkRulesetRuleForbids          1000000    1717 ns/op
ok    github.com/ryanfaerman/fsm 14.864s
```

