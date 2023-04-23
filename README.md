# go-mask

[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/showa-93/go-mask)
![Go Tests](https://github.com/coredns/coredns/actions/workflows/go.test.yml/badge.svg)
[![codecov](https://codecov.io/gh/showa-93/go-mask/branch/main/graph/badge.svg)](https://codecov.io/gh/showa-93/go-mask)
[![Go Report Card](https://goreportcard.com/badge/github.com/showa-93/go-mask)](https://goreportcard.com/report/github.com/showa-93/go-mask)

go-mask is a simple, customizable Go library for masking sensitive information.

- [go-mask](#go-mask)
	- [Features](#features)
	- [Installation](#installation)
	- [Mask Tags](#mask-tags)
	- [How to use](#how-to-use)
		- [string](#string)
		- [int / float64 / uint](#int--float64--uint)
		- [slice / array](#slice--array)
		- [map](#map)
		- [JSON](#json)
		- [nested struct](#nested-struct)
		- [field name / map key](#field-name--map-key)
		- [custom mask function](#custom-mask-function)

## Features

- You can mask any field of a structure using the struct's tags. (example → [How to use](#how-to-use))
- It is also possible to mask using field names or map keys without using tags. (example → [field name / map key](#field-name--map-key))
- Users can make use of their own custom-created masking functions. (example → [custom mask function](#custom-mask-function))
- The masked object is a copied object, so it does not overwrite the original data before masking(although it's not perfect...)
  - Private fields are not copied
  - It is moderately fast in performing deep copies.

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
| mask:"filledXXX" | string | XXX = number of masking characters. Masks with a fixed number of characters. `mask:"filled3"`→`***` |
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
		masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)
		masker.RegisterMaskStringFunc(mask.MaskTypeFixed, masker.MaskFixedString)

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

### int / float64 / uint

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
		masker.RegisterMaskIntFunc(mask.MaskTypeRandom, masker.MaskRandomInt)
		masker.RegisterMaskFloat64Func(mask.MaskTypeRandom, masker.MaskRandomFloat64)

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

### slice / array

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
	masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)
	masker.RegisterMaskIntFunc(mask.MaskTypeRandom, masker.MaskRandomInt)

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
	masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)
	masker.RegisterMaskIntFunc(mask.MaskTypeRandom, masker.MaskRandomInt)

	maskValues, _ := mask.Mask(values)
	maskValues2, _ := masker.Mask(values)

	fmt.Printf("%+v\n", values)
	fmt.Printf("%+v\n", maskValues)
	fmt.Printf("%+v\n", maskValues2)
}
```
```
map[one:{Name:Thomas Jeffrey "Tom" Hanks Type:1} two:{Name:Leonardo Wilhelm DiCaprio Type:2}]
map[one:{Name:************************** Type:8} two:{Name:************************* Type:6}]
map[one:{Name: Type:6} two:{Name: Type:2}]
```

### JSON

```go
package main

import (
	"encoding/json"
	"fmt"

	mask "github.com/showa-93/go-mask"
)

func main() {
	masker := mask.NewMasker()
	masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)
	masker.RegisterMaskField("S", "filled4")

	v := `{
		"S": "Hello world",
		"I": 1,
		"O": {
			"S": "Second",
			"S2": "豚汁"
		}
	}`
	var target any
	json.Unmarshal([]byte(v), &target)
	masked, _ := masker.Mask(target)
	mv, _ := json.Marshal(masked)
	fmt.Println(string(mv))
}
```
```
{"I":1,"O":{"S":"****","S2":"豚汁"},"S":"****"}
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
	masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)

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

### field name / map key

```go
package main

import (
	"fmt"

	mask "github.com/showa-93/go-mask"
)

type User struct {
	ID      string // no tag
	Name    string
	Gender  string
	Age     int
	ExtData map[string]string
}

func main() {
	masker := mask.NewMasker()

	masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)
	masker.RegisterMaskIntFunc(mask.MaskTypeRandom, masker.MaskRandomInt)

	// registered field name
	masker.RegisterMaskField("Name", "filled4")
	masker.RegisterMaskField("Animal", "filled6")
	masker.RegisterMaskField("Age", mask.MaskTypeRandom+"100")

	u := User{
		ID:     "1",
		Name:   "タマ",
		Gender: "Male",
		Age:    4,
		ExtData: map[string]string{
			"Animal": "Cat",
		},
	}
	maskedUser, _ := masker.Mask(u)
	fmt.Printf("%+v", maskedUser)
}
```
```
{ID:1 Name:**** Gender:Male Age:10 ExtData:map[Animal:******]}
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