# glogrotation

Advanged Go! Log Rotation library and CLI implementation.

## Installation

```bash
go get -u github.com/moisespsena-go/glogrotation
```

## Usage

The `Rotator` struct is the concurrent `io.Writer` implementation.

**import library:**

```go
import "github.com/moisespsena-go/glogrotation"
```

### Rotator object

Basic:

```go
rotator := glogrotation.New("app.log")
```

With custom options:

```go
rotator := glogrotation.New("app.log", glogrotation.Options{})
```

#### Options

The `Options` structure:
```go
type Options struct {
	HistoryPath  string // path inside root
	MaxSize      int64 // -1 disables limitter
	Duration     RotationDuration // Default is glogrotation.MaxSize (10M)
	FileMode     os.FileMode
	DirMode      os.FileMode
	HistoryDir   string
	HistoryCount int
}
```

`Options.Duration` field values:
* `glogrotation.Monthly` or char `'M'`
* `glogrotation.Weekly`or char `'W'`
* `glogrotation.Daily`or char `'D'`
* `glogrotation.Hourly` or char `'h'`
* `glogrotation.Minutely` or char `'m'`
* `glogrotation.Yearly` or char `'Y'` 

Load Options From config file:

```go
var opt glogrotation.Options
if f, err := os.Open("options.yml"); err == nil {
	var cfg glogrotation.Config
    if err = yaml.NewDecoder(f).Decode(&cfg); err != nil {
    	panic(err)
    }
    if opt, err = cfg.Options(); err != nil {
    	panic(err)
    }
} else {
	panic(err)
}

rotator := glogrotation.New("app.log", opt)
```

### Examples

Direct write:

```go
n, err := rotator.Write([]byte(`message\n`)) 
```

using `io.Copy`:

```go
var r io.Reader = os.Stdin // your reader
err := io.Copy(rotator, r) 
```


### Implementation example
See to [cli source code](https://github.com/moisespsena-go/glogrotation-cli) for example.

## Author
[Moises P. Sena](https://github.com/moisespsena)
