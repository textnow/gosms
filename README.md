# gosms

This library is for SMS splitting in go.

## Features:
* Smart message splitting
  * Splitting is performed around spaces or after punctuation so that messages remain coherent if concatenation fails at the client.  
* Support for 1 or 2 byte reference numbers in user data headers
* Easily extensible character encoding
  * Comes with support for GSM and UTF-16 character encodings
  * Encodings can be added by implementing the `Encoder` interface

## Usage Example
```
package main

import (
    "fmt"
    "github.com/textnow/gosms"
)

func main() {
    var (
        from         = "from"
        to           = []string{"to"}
        message      = "This message should be split depending on the placement of spaces and " +
                       "punctuation. If the client fails to stitch the message segments back " +
                       "together, the user should still be able to read this text."
        messageBytes = 55
    )

	splitter := gosms.NewSplitter()
	splitter.SetMessageBytes(messageBytes)

    SMSs, _ := splitter.Split(from, to, message)

    for idx, sms := range SMSs {
        fmt.Printf("SMS #%d\n", idx + 1)
        fmt.Printf("from    : %s\n", sms.GetFrom())
        fmt.Printf("to      : %s\n", sms.GetTo())
        fmt.Printf("content : \"%s\"\n", sms.GetContent())
        fmt.Println()
    }
}
```
## Output
```
SMS #1
from    : from
to      : to
content : "This message should be split depending on the placement "

SMS #2
from    : from
to      : to
content : "of spaces and punctuation. If the client fails to stitch"

SMS #3
from    : from
to      : to
content : " the message segments back together, the user should "

SMS #4
from    : from
to      : to
content : "still be able to read this text."

```
