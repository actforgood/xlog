# Xlog

[![Build Status](https://github.com/actforgood/xlog/actions/workflows/build.yml/badge.svg)](https://github.com/actforgood/xlog/actions/workflows/build.yml)
[![License](https://img.shields.io/badge/license-MIT-blue)](https://raw.githubusercontent.com/actforgood/xlog/main/LICENSE)
[![Coverage Status](https://coveralls.io/repos/github/actforgood/xlog/badge.svg?branch=main)](https://coveralls.io/github/actforgood/xlog?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/actforgood/xlog.svg)](https://pkg.go.dev/github.com/actforgood/xlog)  

---  

Package `xlog` provides a structured leveled Logger implemented in two different strategies: synchronous and asynchronous.  
The logs can be formatted in JSON, logfmt, custom text format.  


### Installation

```shell
$ go get -u github.com/actforgood/xlog
```


### Supported level APIs:
* Critical
* Error
* Warning
* Info
* Debug
* Log // arbitrary log


### Common options
A logger will need a `CommonOpts` through which you can configure some default keys and values used by the logger.
```go
xOpts := xlog.NewCommonOpts() // instantiates new CommonOpts object with default values.
``` 
###### Configuring `level` options for a log.  
Example of applicability:  

* you may want to log from Error above messages on production env, and all levels on dev env.  
* you may want to log messages by different level in different sources - see also `ExampleMultiLogger_splitMessagesByLevel` from doc reference.  

```go
xOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelDebug) // by default is LevelWarning
xOpts.MaxLevel = xlog.FixedLevelProvider(xlog.LevelInfo) // by default is LevelCritical
xOpts.LevelKey = "level" // by default is "lvl"
xOpts.LevelLabels = map[xlog.Level]string{ // by default "CRITICAL", "ERROR", "WARN", "INFO", "DEBUG" are used
    xlog.LevelCritical: "CRT",
    xlog.LevelError: "ERR",
    xlog.LevelWarning: "WRN",
    xlog.LevelInfo: "INF",
    xlog.LevelDebug: "DBG", 
}
```  
Check also the `xlog.EnvLevelProvider` - to get the level from OS's env.  
You can make your own `xlog.LevelProvider` - to get the level from a remote API/other source, for example.  

###### Configuring `time` options for a log.
```go
xOpts.Time    = xlog.UTCTimeProvider(time.RFC3339) // by default is time.RFC3339Nano
xOpts.TimeKey = "t" // by default is "date"
```
Check also the `xlog.LocalTimeProvider` - to get time in local server timezone.  
You can make your own `xlog.Provider` if needed for more custom logic.  

###### Configuring `source` options for a log.
```go
xOpts.Source = xlog.SourceProvider(4, 2) // by default logs full path with a stack level of 4.
xOpts.SourceKey = "source" // by default is "src"
```
By setting `SourceKey` to blank, you can disable source logging.
By changing the first parameter in `SourceProvider`, you can manipulate the level in the stack trace.
By changing the second parameter in `SourceProvider`, you can manipulate how many levels in the path to be logged.
Example:
```go
xlog.SourceProvider(4, 0) // => "src":"/Users/JohnDoe/work/go/xlog/example.go:65" (full path)
xlog.SourceProvider(4, 1) // => "src":"/example.go:65"
xlog.SourceProvider(4, 2) // => "src":"/xlog/example.go:65"
xlog.SourceProvider(4, 3) // => "src":"/go/xlog/example.go:65"
...
```

###### Configuring additional key-values to be logged with every log.
```go
xOpts.AdditionalKeyValues = []interface{}{
	"app", "demoXlog",
	"env", "prod",
	"release", "v1.10.0",
}
```

###### Configuring an I/O / formatting error handler for errors that may occur during logging.
By design, logger contract does not return error from its methods.
A no operation `ErrorHandler` is set by default. You can change it to something else
if suitable. For example, log with standard go logger the error.
```go
xOpts.ErrHandler = func(err error, keyValues []interface{}) {
	// import "log"
	log.Printf("An error occurred during logging. err = %v, logParams = %v", err, keyValues)
}
```


### Loggers

##### SyncLogger
`SyncLogger` is a `Logger` which writes logs synchronously.  
It just calls underlying writer with each log call.  
Note: if used in a concurrent context, log writes are not concurrent safe, unless the writer is concurrent safe. See also `NewSyncWriter` on this matter.  
Example of usage:
```go
xLogger := xlog.NewSyncLogger(os.Stdout)
defer xLogger.Close()
xLogger.Error(
	xlog.MessageKey, "Could not read file",
	xlog.ErrorKey, io.ErrUnexpectedEOF,
	"file", "/some/file",
)
```
You can change the formatter (json is default), and common options like:
```go
xOpts := xlog.NewCommonOpts()
xOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelInfo)
xLogger := xlog.NewSyncLogger(
	os.Stdout,
	xlog.SyncLoggerWithOptions(xOpts),
	xlog.SyncLoggerWithFormatter(xlog.LogfmtFormatter),
)
defer xLogger.Close()
```

##### AsyncLogger
`AsyncLogger` is a `Logger` which writes logs asynchronously.  
Note: if used in a concurrent context, log writes are concurrent safe if only one worker is configured to process the logs. Otherwise, log writes are not concurrent safe, unless the writer is concurrent safe. See also `NewSyncWriter` and `AsyncLoggerWithWorkersNo` on this matter.  
Example of usage:
```go
xLogger := xlog.NewAsyncLogger(os.Stdout)
defer xLogger.Close()
xLogger.Error(
	xlog.MessageKey, "Could not read file",
	xlog.ErrorKey, io.ErrUnexpectedEOF,
	"file", "/some/file",
)
```
You can change some options on it like:
```go
xOpts := xlog.NewCommonOpts()
xOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelInfo)
xLogger := xlog.NewAsyncLogger(
	os.Stdout,
	xlog.AsyncLoggerWithOptions(xOpts),
	xlog.AsyncLoggerWithFormatter(xlog.LogfmtFormatter),     // defaults to json
	xlog.AsyncLoggerWithWorkersNo(uint16(runtime.NumCPU())), // defaults to 1
	xlog.AsyncLoggerWithChannelSize(512),                    // defaults to 256
)
defer xLogger.Close()
```

###### Benchmark example between sync / async loggers
```
go test -run=^# -benchmem -benchtime=5s -bench ".*(sequential|parallel)"
goos: darwin
goarch: amd64
pkg: github.com/actforgood/xlog
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkAsyncLogger_json_withDiscardWriter_with256ChanSize_with1Worker_sequential-8     1855483              3241 ns/op            1704 B/op         32 allocs/op
BenchmarkAsyncLogger_json_withDiscardWriter_with256ChanSize_with1Worker_parallel-8       1713628              3565 ns/op            1704 B/op         32 allocs/op
BenchmarkSyncLogger_json_withDiscardWriter_sequential-8                                   984081              5269 ns/op            1696 B/op         32 allocs/op
BenchmarkSyncLogger_json_withDiscardWriter_parallel-8                                    3394920              1797 ns/op            1696 B/op         32 allocs/op
```
Note how in a high concurrency context (*_parallel*) the sync logger actually behaves more well than async one.

##### MultiLogger
`MultiLogger` is a composite `Logger` capable of logging to multiple loggers.  
Example of usage:
```go
xLogger := xlog.NewMultiLogger(loggerA, loggerB)
defer xLogger.Close()
xLogger.Error(
	xlog.MessageKey, "Could not read file",
	xlog.ErrorKey, io.ErrUnexpectedEOF,
	"file", "/some/file",
)
```

##### NopLogger
`NopLogger` is a no-operation `Logger` which does nothing. It simply ignores any log.  
You can use it when benchmarking another component that uses logger, for example, in order for the logging process not to interfere with the main component's bench stats.

##### MockLogger
`MockLogger` is a mock for `Logger` contract, to be used in Unit Tests.


### Formats

##### JSONFormatter
Logs get written in JSON format. Is the default format configured for sync / async loggers.  
Example of log:
```javascript
{"appName":"demo","date":"2022-03-16T16:01:20Z","env":"dev","lvl":"DEBUG","msg":"Hello World","src":"/logger_async_test.go:43","year":2022}
```

##### LogfmtFormatter
Logs get written in [logfmt](https://brandur.org/logfmt) format.  
Example of configuring:  
```go
xLogger := xlog.NewSyncLogger(
	os.Stdout,
	xlog.SyncLoggerWithOptions(xOpts),
	xlog.SyncLoggerWithFormatter(xlog.LogfmtFormatter),
)
```

Example of log:  
```
date=2022-04-12T16:01:20Z lvl=INFO src=/formatter_logfmt_test.go:42 appName=demo env=dev msg="Hello World" year=2022
```

##### TextFormatter
Logs get written in custom, human friendly format: *TIME SOURCE LEVEL MESSAGE KEY1=VALUE1 KEY2=VALUE2 ...*  
Note: this is not a structured logging format. It can be used for a "dev" logger, for example.  
Example of configuring (see also `ExampleSyncLogger_devLogger` from doc reference):  
```go
xLogger := xlog.NewSyncLogger(
	os.Stdout,
	xlog.SyncLoggerWithOptions(xOpts),
	xlog.SyncLoggerWithFormatter(xlog.TextFormatter(xOpts)),
)
```

Example of log:  
```
2022-03-14T16:01:20Z /formatter_text_test.go:40 DEBUG Hello World year=2022
```

##### SyslogFormatter
Logs get written to system syslog.
Example of configuring (see also `ExampleSyncLogger_withSyslog` from doc reference):
```go
xLogger := xlog.NewSyncLogger(
	syslogWriter,
	xlog.SyncLoggerWithFormatter(xlog.SyslogFormatter(
		xlog.JSONFormatter,
		xlog.NewDefaultSyslogLevelProvider(xOpts),
		"",
	)),
	xlog.SyncLoggerWithOptions(xOpts),
)
```

##### SentryFormatter
Logs get written to [Sentry](https://docs.sentry.io/).
Example of configuring (see also `ExampleSyncLogger_withSentry` from doc reference):
```go
xLogger := xlog.NewSyncLogger(
	io.Discard, // no need for other writer, SentryFormatter will override it with a buffered one in order to get original Formatter output.
	xlog.SyncLoggerWithOptions(xOpts),
	xlog.SyncLoggerWithFormatter(xlog.SentryFormatter(
		xlog.JSONFormatter,
		sentry.CurrentHub().Clone(), // make a clone if you're not using sentry only in the logger.
		xOpts,
	)),
)
```


### Writers

##### SyncWriter
`SyncWriter` decorates an `io.Writer` so that each call to Write is synchronized with a mutex, making is safe for concurrent use by multiple goroutines.  
It should be used if writer's `Write` method is not thread safe.  
For example an `os.File` is safe, so it doesn't need this wrapper, on the other hand, a `bytes.Buffer` is not.  

##### BufferedWriter
`BufferedWriter` decorates an `io.Writer` so that written bytes are buffered.  
It is concurrent safe to use.  
It has the capability of auto-flushing the buffer, time interval based. This capability can also be disabled.
If an error occurs in the write process, at next log write, this error is not persisted, opposite using directly a `bufio.Writer` (see [this](https://github.com/golang/go/blob/go1.17.3/src/bufio/bufio.go#L633)).  
Example of benchmarks between directly writes to a file, and writing to a "buffered" file:
```
go test -run=^# -benchmem -benchtime=5s -bench ".*FileWriter"
goos: darwin
goarch: amd64
pkg: github.com/actforgood/xlog
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkAsyncLogger_json_withFileWriter_with256ChanSize_with1Worker-8                            
666721               9007 ns/op            1704 B/op           32 allocs/op
BenchmarkAsyncLogger_json_withBufferedFileWriter_with256ChanSize_with1Worker-8                   
1597966              3696 ns/op            1704 B/op           32 allocs/op

BenchmarkSyncLogger_json_withFileWriter-8                                                         
507146             10856 ns/op            1696 B/op           32 allocs/op
BenchmarkSyncLogger_json_withBufferedFileWriter-8                                                 
920844              5928 ns/op            1696 B/op           32 allocs/op
```


### Misc 
Feel free to use this logger if you like it and fits your needs.  
Check also other popular, performant loggers like Uber Zap, Zerolog, Gokit...  
Here stands some benchmarks made locally based on [this](https://github.com/imkira/go-loggers-bench) repo.  
```
go test -run=^# -benchmem -benchtime=5s -bench ".*JSON"
goos: darwin
goarch: amd64
pkg: github.com/imkira/go-loggers-bench
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkGokitJSONPositive-8          3904477       1466 ns/op      1544 B/op      24 allocs/op
BenchmarkLog15JSONPositive-8           973974       5464 ns/op      2009 B/op      30 allocs/op
BenchmarkLogrusJSONPositive-8         2950423       1986 ns/op      2212 B/op      34 allocs/op
BenchmarkXlogSyncJSONPositive-8       3880016       1520 ns/op      1662 B/op      28 allocs/op
BenchmarkZerologJSONPositive-8       28168381      202.8 ns/op         0 B/op       0 allocs/op

BenchmarkGokitJSONNegative-8        180957508      32.65 ns/op       128 B/op       1 allocs/op
BenchmarkLog15JSONNegative-8         12070347      466.3 ns/op       632 B/op       5 allocs/op
BenchmarkLogrusJSONNegative-8        24485853      211.4 ns/op       496 B/op       4 allocs/op
BenchmarkXlogSyncJSONNegative-8    1000000000      3.415 ns/op         0 B/op       0 allocs/op
BenchmarkZerologJSONNegative-8     1000000000      3.288 ns/op         0 B/op       0 allocs/op
```


### License
This package is released under a MIT license. See [LICENSE](LICENSE).  
Other 3rd party packages directly used by this package are released under their own licenses.  

* github.com/getsentry/sentry-go - [BSD 2 Clause](https://github.com/getsentry/sentry-go/blob/master/LICENSE)  
* github.com/go-logfmt/logfmt - [MIT License](https://github.com/go-logfmt/logfmt/blob/main/LICENSE)  
* github.com/actforgood/xerr - [MIT License](https://github.com/actforgood/xerr/blob/main/LICENSE)  
