package jog

// Middleware defines how a middleware will functions.
type Middleware func(handler ExecutionHandler) ExecutionHandler
