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
		masker := mask.NewMasker()
		masker.SetMaskChar("-")

		maskValue, _ := mask.Mask(value)
		maskValue2, _ := masker.Mask(value)

		fmt.Printf("%+v\n", value)
		fmt.Printf("%+v\n", maskValue)
		fmt.Printf("%+v\n", maskValue2)
	}
}
```
```
********
{Title:Catch Me If You Can Casts:[Thomas Jeffrey "Tom" Hanks Leonardo Wilhelm DiCaprio]}
{Title:******************* Casts:[******** ********]}
{Title:------------------- Casts:[-------- --------]}
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
		masker := mask.NewMasker()

		maskValue, _ := mask.Mask(value)
		maskValue2, _ := masker.Mask(value)

		fmt.Printf("%+v\n", maskValue)
		fmt.Printf("%+v\n", maskValue2)
	}
}
```
```
29
50.45
{Price:917 Percent:0.183}
{Price:733 Percent:0.241}
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
	masker := mask.NewMasker()
	masker.SetMaskChar("+")

	maskValues, _ := mask.Mask(values)
	maskValues2, _ := masker.Mask(values)

	fmt.Printf("%+v\n", values)
	fmt.Printf("%+v\n", maskValues)
	fmt.Printf("%+v\n", maskValues2)
}
```
```
[{Name:Thomas Jeffrey "Tom" Hanks Type:1} {Name:Leonardo Wilhelm DiCaprio Type:2}]
[{Name:************************** Type:8} {Name:************************* Type:9}]
[{Name:++++++++++++++++++++++++++ Type:4} {Name:+++++++++++++++++++++++++ Type:8}]
```

### map

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
	values := map[string]Value{
		"one": {
			Name: "Thomas Jeffrey \"Tom\" Hanks",
			Type: 1,
		},
		"two": {
			Name: "Leonardo Wilhelm DiCaprio",
			Type: 2,
		},
	}
	masker := mask.NewMasker()
	masker.SetMaskChar("")

	maskValues, _ := mask.Mask(values)
	maskValues2, _ := masker.Mask(values)

	fmt.Printf("%+v\n", values)
	fmt.Printf("%+v\n", maskValues)
	fmt.Printf("%+v\n", maskValues2)
}
```
```
map[one:{Name:Thomas Jeffrey "Tom" Hanks Type:1} two:{Name:Leonardo Wilhelm DiCaprio Type:2}]
map[one:{Name:************************** Type:4} two:{Name:************************* Type:8}]
map[one:{Name:++++++++++++++++++++++++++ Type:2} two:{Name:+++++++++++++++++++++++++ Type:3}]
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

	masker := mask.NewMasker()
	masker.SetMaskChar("🤗")

	maskNode, _ := mask.Mask(node)
	maskNode2, _ := masker.Mask(node)

	fmt.Printf("first=%+v,second=%+v,third=%+v\n", node, node.Next, node.Next.Next)
	fmt.Printf("first=%+v,second=%+v,third=%+v\n", maskNode, maskNode.Next, maskNode.Next.Next)
	fmt.Printf("first=%+v,second=%+v,third=%+v\n", maskNode2.(Node), maskNode2.(Node).Next, maskNode2.(Node).Next.Next)
}
```
```
first={Value:first Next:0xc000010048},second=&{Value:second Next:0xc000010060},third=&{Value:third Next:<nil>}
first={Value:***** Next:0xc0000100a8},second=&{Value:****** Next:0xc0000100c0},third=&{Value:***** Next:<nil>}
first={Value:🤗🤗🤗🤗🤗 Next:0xc000010120},second=&{Value:🤗🤗🤗🤗🤗🤗 Next:0xc000010138},third=&{Value:🤗🤗🤗🤗🤗 Next:<nil>}
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

	// The Masker initialized with NewMasker does not have
	// any custom masking functions registered, so no masking will occur
	masker := mask.NewMasker()
	got2, _ := masker.Mask(input)
	fmt.Printf("%+v\n", got2)
}

```
```
{Message:I love gopher!}
{Message:I love cat!}
{Message:I love gopher!}
```