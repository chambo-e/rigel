# Universally Unique Lexicographically Sortable Identifier

![Project status](https://img.shields.io/badge/version-0.1.0-yellow.svg)
[![Build Status](https://secure.travis-ci.org/oklog/ulid.png)](http://travis-ci.org/oklog/ulid)
[![Go Report Card](https://goreportcard.com/badge/oklog/ulid?cache=0)](https://goreportcard.com/report/oklog/ulid)
[![Coverage Status](https://coveralls.io/repos/github/oklog/ulid/badge.svg?branch=master&cache=0)](https://coveralls.io/github/oklog/ulid?branch=master)
[![GoDoc](https://godoc.org/github.com/oklog/ulid?status.svg)](https://godoc.org/github.com/oklog/ulid)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/oklog/ulid/master/LICENSE)

A Go port of [alizain/ulid](https://github.com/alizain/ulid) with binary format implemented.

## Background

A GUID/UUID can be suboptimal for many use-cases because:

- It isn't the most character efficient way of encoding 128 bits
- It provides no other information than randomness

A ULID however:

- Is compatible with UUID/GUID's
- 1.21e+24 unique ULIDs per millisecond (1,208,925,819,614,629,174,706,176 to be exact)
- Lexicographically sortable
- Canonically encoded as a 26 character string, as opposed to the 36 character UUID
- Uses Crockford's base32 for better efficiency and readability (5 bits per character)
- Case insensitive
- No special characters (URL safe)

## Install

```shell
go get github.com/oklog/ulid
```

## Usage

An ULID is constructed with a `time.Time` and an `io.Reader` entropy source.
This design allows for greater flexibility in choosing your trade-offs.

Please note that `rand.Rand` from the `math` package is *not* safe for concurrent use.
Instantiate one per long living go-routine or use a `sync.Pool` if you want to avoid the potential contention of a locked `rand.Source` as its been frequently observed in the package level functions.

```go
func ExampleULID() {
	t := time.Unix(1000000, 0)
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	fmt.Println(ulid.MustNew(ulid.Timestamp(t), entropy))
	// Output: 0000XSNJG0MQJHBF4QX1EFD6Y3
}

```

## Specification

Below is the current specification of ULID as implemented in this repository.

### Components

**Timestamp**
- 48 bits
- UNIX-time in milliseconds
- Won't run out of space till the year 10895 AD

**Entropy**
- 80 bits
- User defined entropy source.

### Encoding

[Crockford's Base32](http://www.crockford.com/wrmg/base32.html) is used as shown.
This alphabet excludes the letters I, L, O, and U to avoid confusion and abuse.

```
0123456789ABCDEFGHJKMNPQRSTVWXYZ
```

### Binary Layout and Byte Order

The components are encoded as 16 octets. Each component is encoded with the Most Significant Byte first (network byte order).

```
0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      32_bit_uint_time_high                    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     16_bit_uint_time_low      |       16_bit_uint_random      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       32_bit_uint_random                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       32_bit_uint_random                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### String Representation

```
 01AN4Z07BY      79KA1307SR9X4MV3
|----------|    |----------------|
 Timestamp           Entropy
  10 chars           16 chars
   48bits             80bits
   base32             base32
```

## Test

```shell
go test ./...
```

## Benchmarks

On a Intel Core i7 Ivy Bridge 2.7 GHz, MacOS 10.12.1 and Go 1.8.0beta1

```
BenchmarkNew/WithEntropy-8          20000000      62.6 ns/op    16 B/op    1 allocs/op
BenchmarkNew/WithoutEntropy-8       50000000      29.6 ns/op    16 B/op    1 allocs/op
BenchmarkMustNew/WithEntropy-8      20000000      67.4 ns/op    16 B/op    1 allocs/op
BenchmarkMustNew/WithoutEntropy-8   50000000      33.8 ns/op    16 B/op    1 allocs/op
BenchmarkParse-8                    50000000      29.8 ns/op     0 B/op    0 allocs/op
BenchmarkMustParse-8                50000000      34.6 ns/op     0 B/op    0 allocs/op
BenchmarkString-8                   20000000      61.4 ns/op    32 B/op    1 allocs/op
BenchmarkMarshal/Text-8             30000000      52.4 ns/op    32 B/op    1 allocs/op
BenchmarkMarshal/TextTo-8           100000000     22.5 ns/op     0 B/op    0 allocs/op
BenchmarkMarshal/Binary-8           300000000     4.15 ns/op     0 B/op    0 allocs/op
BenchmarkMarshal/BinaryTo-8         2000000000    1.18 ns/op     0 B/op    0 allocs/op
BenchmarkUnmarshal/Text-8           100000000     20.6 ns/op     0 B/op    0 allocs/op
BenchmarkUnmarshal/Binary-8         300000000     4.88 ns/op     0 B/op    0 allocs/op
BenchmarkNow-8                      50000000      37.6 ns/op     0 B/op    0 allocs/op
BenchmarkTimestamp-8                50000000      24.9 ns/op     0 B/op    0 allocs/op
BenchmarkTime-8                     2000000000    0.57 ns/op     0 B/op    0 allocs/op
BenchmarkSetTime-8                  2000000000    0.84 ns/op     0 B/op    0 allocs/op
BenchmarkEntropy-8                  200000000     7.44 ns/op     0 B/op    0 allocs/op
BenchmarkSetEntropy-8               2000000000    0.86 ns/op     0 B/op    0 allocs/op
```

## Prior Art

- [alizain/ulid](https://github.com/alizain/ulid)
- [RobThree/NUlid](https://github.com/RobThree/NUlid)
- [imdario/go-ulid](https://github.com/imdario/go-ulid)
