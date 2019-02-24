jsonquery
====
[![Build Status](https://travis-ci.org/antchfx/jsonquery.svg?branch=master)](https://travis-ci.org/antchfx/jsonquery)
[![Coverage Status](https://coveralls.io/repos/github/antchfx/jsonquery/badge.svg?branch=master)](https://coveralls.io/github/antchfx/jsonquery?branch=master)
[![GoDoc](https://godoc.org/github.com/antchfx/jsonquery?status.svg)](https://godoc.org/github.com/antchfx/jsonquery)
[![Go Report Card](https://goreportcard.com/badge/github.com/antchfx/jsonquery)](https://goreportcard.com/report/github.com/antchfx/jsonquery)

Overview
===

jsonquery is an XPath query package for JSON document, lets you extract data from JSON documents through an XPath expression.

Getting Started
===

### Install Package

> $ go get github.com/antchfx/jsonquery

#### Load JSON document from URL.

```go
doc, err := jsonquery.LoadURL("http://www.example.com/feed?json")
```

#### Load JSON document from string.

```go
s :=`{
    "name":"John",
    "age":31, 
    "city":"New York" 
    }`
doc, err := jsonquery.Parse(strings.NewReader(s))
```

#### Load JSON document from io.Reader.

```go
f, err := os.Open("./books.json")
doc, err := jsonquery.Parse(f)
```

#### Find authors of all books in the store.
```go
list := jsonquery.Find(doc, "store/book/*/author")
// or equal to
list := jsonquery.Find(doc, "//author")
```

#### Find the third book.

```go
book := jsonquery.Find(doc, "//book/*[3]")
```

#### Find the last book.

```go
book := jsonquery.Find(doc, "//book/*[last()]")
```

#### Find all books with isbn number.

```go
list := jsonquery.Find(doc, "//book/*[isbn]")
```

#### Find all books cheapier than 10.

```go
list := jsonquery.Find(doc, "//book/*[price<10]")
```

Quick Tutorial
===

```go
func main() {
	s := `{
		"name": "John",
		"age"      : 26,
		"address"  : {
		  "streetAddress": "naist street",
		  "city"         : "Nara",
		  "postalCode"   : "630-0192"
		},
		"phoneNumbers": [
		  {
			"type"  : "iPhone",
			"number": "0123-4567-8888"
		  },
		  {
			"type"  : "home",
			"number": "0123-4567-8910"
		  }
		]
	}`
	doc, _ := jsonquery.Parse(strings.NewReader(s))
	name := jsonquery.FindOne(doc, "name")
	fmt.Printf("name: %s\n", name.InnerText())
	var a []string
	for _, n := range jsonquery.Find(doc, "phoneNumbers/*/number") {
		a = append(a, n.InnerText())
	}
	fmt.Printf("phone number: %s\n", strings.Join(a, ","))
	if n := jsonquery.FindOne(doc, "address/streetAddress"); n != nil {
		fmt.Printf("address: %s\n", n.InnerText())
	}
}
```

Implement Principle
===
If you are familiar with XPath and XML, you can quick start to known how 
write your XPath expression.

```json
{
"name":"John",
"age":30,
"cars": [
	{ "name":"Ford", "models":[ "Fiesta", "Focus", "Mustang" ] },
	{ "name":"BMW", "models":[ "320", "X3", "X5" ] },
	{ "name":"Fiat", "models":[ "500", "Panda" ] }
]
}
```
The above JSON document will be convert to similar to XML document by the *JSONQuery*, like below:

```XML
<name>John</name>
<age>30</age>
<cars>
	<element>
		<name>Ford</name>
		<models>
			<element>Fiesta</element>
			<element>Focus</element>
			<element>Mustang</element>
		</models>		
	</element>
	<element>
		<name>BMW</name>
		<models>
			<element>320</element>
			<element>X3</element>
			<element>X5</element>
		</models>		
	</element>
	<element>
		<name>Fiat</name>
		<models>
			<element>500</element>
			<element>Panda</element>
		</models>		
	</element>
</cars>
```

Notes: `element` is empty element that have no any name.

List of supported XPath query packages
===
|Name |Description |
|--------------------------|----------------|
|[htmlquery](https://github.com/antchfx/htmlquery) | XPath query package for the HTML document|
|[xmlquery](https://github.com/antchfx/xmlquery) | XPath query package for the XML document|
|[jsonquery](https://github.com/antchfx/jsonquery) | XPath query package for the JSON document|
