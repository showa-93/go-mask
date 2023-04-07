# go-mask

[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/showa-93/go-mask)
![Go Tests](https://github.com/coredns/coredns/actions/workflows/go.test.yml/badge.svg)
[![codecov](https://codecov.io/gh/showa-93/go-mask/branch/main/graph/badge.svg)](https://codecov.io/gh/showa-93/go-mask)
[![Go Report Card](https://goreportcard.com/badge/github.com/showa-93/go-mask)](https://goreportcard.com/report/github.com/showa-93/go-mask)

go-mask is a simple, customizable Go library for masking sensitive information.

## Features

- You can mask any field of a structure using the struct's tags
- Users can make use of their own custom-created masking functions
- The masked object is a copied object, so it does not overwrite the original data before masking(although it's not perfect...)
  - Private fields are not copied

## Installation

```sh
go get github.com/showa-93/go-mask
```

## Mask Tags

go-mask does not provide many tags by default.  
This is because it is believed that users should create their own necessary masking functions.  

| tag | type | description |
| :-- | :-- | :-- |
| mask:"filled" | string | Masks the string with the same number of masking characters. |
| mask:"fixed" | string | Masks with a fixed number of characters. `*******` |
| mask:"hash" | string | Masks the string by converting it to a value using sha1. |
| mask:"randomXXX" | int / float64 | XXX = numeric value. Masks with a random value in the range of 0 to the XXX. |
| mask:"zero" | any | It can be applied to any type, masking it with the zero value of that type. |

## How to use

### string

```go
package main

import (
	"fmt"

	mask "github.com/showa-93/go-mask"
)

func main() {
	{
		maskValue, _ := mask.String(mask.MaskTypeFixed, "Hello World!!")
		fmt.Println(maskValue)
	}
	{
		value := struct {
			Title string   `mask:"filled"`
			Casts []string `mask:"fixed"`
		}{
			Title: "Catch Me If You Can",
			Casts: []string{
				"Thomas Jeffrey \"Tom\" Hanks",
				"Leonardo Wilhelm DiCaprio",
			},
		}
		maskValue, _ := mask.Mask(value)
		fmt.Printf("%+v\n", value)
		fmt.Printf("%+v\n", maskValue)
	}
}
```
```
********
{Title:Catch Me If You Can Casts:[Thomas Jeffrey "Tom" Hanks Leonardo Wilhelm DiCaprio]}
{Title:******************* Casts:[******** ********]}
```

### int / float64

```go
package main

import (
	"fmt"

	mask "github.com/showa-93/go-mask"
)

func main() {
	{
		maskValue, _ := mask.Int("random100", 10)
		fmt.Println(maskValue)
	}
	{
		maskValue, _ := mask.Float64("random100.2", 12.3)
		fmt.Println(maskValue)
	}

	{
		value := struct {
			Price   int     `mask:"random1000"`
			Percent float64 `mask:"random1.3"`
		}{
			Price:   300,
			Percent: 0.80,
		}
		maskValue, _ := mask.Mask(value)
		fmt.Printf("%+v\n", maskValue)
	}
}

```
```
70
85.12
{Price:375 Percent:0.653}
```

### slice

```go
package main

import (
	"fmt"

	"github.com/showa-93/go-mask"
)

type Value struct {
	Name string `mask:"filled"`
	Type int    `mask:"random10"`
}

func main() {
	values := []Value{
		{
			Name: "Thomas Jeffrey \"Tom\" Hanks",
			Type: 1,
		},
		{
			Name: "Leonardo Wilhelm DiCaprio",
			Type: 2,
		},
	}

	maskValues, _ := mask.Mask(values)
  fmt.Printf("%+v\n", values)
	fmt.Printf("%+v\n", maskValues)
}
```
```
[{Name:Thomas Jeffrey "Tom" Hanks Type:1} {Name:Leonardo Wilhelm DiCaprio Type:2}]
[{Name:************************** Type:7} {Name:************************* Type:0}]
```

### nested struct

```go
package main

import (
	"fmt"

	"github.com/showa-93/go-mask"
)

type Node struct {
	Value string `mask:"filled"`
	Next  *Node
}

func main() {
	node := Node{
		Value: "first",
		Next: &Node{
			Value: "second",
			Next: &Node{
				Value: "third",
			},
		},
	}
	maskNode, _ := mask.Mask(node)
	fmt.Printf("first=%+v,second=%+v,third=%+v\n", node, node.Next, node.Next.Next)
	fmt.Printf("first=%+v,second=%+v,third=%+v\n", maskNode, maskNode.Next, maskNode.Next.Next)
}
```
```
first={Value:first Next:0xc000010048},second=&{Value:second Next:0xc000010060},third=&{Value:third Next:<nil>}
first={Value:***** Next:0xc0000100a8},second=&{Value:****** Next:0xc0000100c0},third=&{Value:***** Next:<nil>}
```

### custom mask function

```go
package main

import (
	"fmt"
	"regexp"
	"strings"

	mask "github.com/showa-93/go-mask"
)

func init() {
	maskTypeRegExp := "regexp"
	mask.RegisterMaskStringFunc(maskTypeRegExp, MaskRegExp)
}

// MaskRegExp is sample to add a custom mask function
func MaskRegExp(arg, value string) (string, error) {
	var (
		reg *regexp.Regexp
		err error
	)
	reg, err = regexp.Compile(arg)
	if err != nil {
		return "", err
	}

	indexes := reg.FindStringSubmatchIndex(value)
	if len(indexes) >= 4 && indexes[2] >= 0 && indexes[3] >= 0 {
		var sb strings.Builder
		sb.WriteString(value[:indexes[2]])
		sb.WriteString(mask.MaskChar())
		sb.WriteString(value[indexes[3]:])
		return sb.String(), nil
	}

	return value, nil
}

func main() {
	mask.SetMaskChar("cat")
	type Hachiware struct {
		Message string `mask:"regexp(gopher)."`
	}

	input := Hachiware{Message: "I love gopher!"}
	got, _ := mask.Mask(input)
	fmt.Printf("%+v\n", input)
	fmt.Printf("%+v\n", got)
}
```
```
{Message:I love gopher!}
{Message:I love cat!}
```