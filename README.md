# wee

## refer

- [Go Doc Comments - The Go Programming Language](https://tip.golang.org/doc/comment)

## Faq

### panic: context canceled

context canceled issue usually happend when elem is bound with Timeout on `bot.Elem(xxx)`, you can use

```go
elem.CancelTimeout().Timeout(xxx) to rewrite previous value.
```
