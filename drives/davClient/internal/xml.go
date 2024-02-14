package internal

import (
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// RawXMLValue is a raw XML value. It implements xml.Unmarshaler and
// xml.Marshaler and can be used to delay XML decoding or precompute an XML
// encoding.
type RawXMLValue struct {
	tok      xml.Token // guaranteed not to be xml.EndElement
	children []RawXMLValue

	// Unfortunately encoding/xml doesn't offer TokenWriter, so we need to
	// cache outgoing data.
	out interface{}
}

// NewRawXMLElement creates a new RawXMLValue for an element.
func NewRawXMLElement(name xml.Name, attr []xml.Attr, children []RawXMLValue) *RawXMLValue {
	return &RawXMLValue{tok: xml.StartElement{name, attr}, children: children}
}

// EncodeRawXMLElement encodes a value into a new RawXMLValue. The XML value
// can only be used for marshalling.
func EncodeRawXMLElement(v interface{}) (*RawXMLValue, error) {
	return &RawXMLValue{out: v}, nil
}

// UnmarshalXML implements xml.Unmarshaler.
func (val *RawXMLValue) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	val.tok = start
	val.children = nil
	val.out = nil

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			child := RawXMLValue{}
			if err := child.UnmarshalXML(d, tok); err != nil {
				return err
			}
			val.children = append(val.children, child)
		case xml.EndElement:
			return nil
		default:
			val.children = append(val.children, RawXMLValue{tok: xml.CopyToken(tok)})
		}
	}
}

// MarshalXML implements xml.Marshaler.
func (val *RawXMLValue) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if val.out != nil {
		return e.Encode(val.out)
	}

	switch tok := val.tok.(type) {
	case xml.StartElement:
		if err := e.EncodeToken(tok); err != nil {
			return err
		}
		for _, child := range val.children {
			// TODO: find a sensible value for the start argument?
			if err := child.MarshalXML(e, xml.StartElement{}); err != nil {
				return err
			}
		}
		return e.EncodeToken(tok.End())
	case xml.EndElement:
		panic("unexpected end element")
	default:
		return e.EncodeToken(tok)
	}
}

var _ xml.Marshaler = (*RawXMLValue)(nil)
var _ xml.Unmarshaler = (*RawXMLValue)(nil)

func (val *RawXMLValue) Decode(v interface{}) error {
	return xml.NewTokenDecoder(val.TokenReader()).Decode(&v)
}

func (val *RawXMLValue) XMLName() (name xml.Name, ok bool) {
	if start, ok := val.tok.(xml.StartElement); ok {
		return start.Name, true
	}
	return xml.Name{}, false
}

// TokenReader returns a stream of tokens for the XML value.
func (val *RawXMLValue) TokenReader() xml.TokenReader {
	if val.out != nil {
		panic("webdav: called RawXMLValue.TokenReader on a marshal-only XML value")
	}
	return &rawXMLValueReader{val: val}
}

type rawXMLValueReader struct {
	val         *RawXMLValue
	start, end  bool
	child       int
	childReader xml.TokenReader
}

func (tr *rawXMLValueReader) Token() (xml.Token, error) {
	if tr.end {
		return nil, io.EOF
	}

	start, ok := tr.val.tok.(xml.StartElement)
	if !ok {
		tr.end = true
		return tr.val.tok, nil
	}

	if !tr.start {
		tr.start = true
		return start, nil
	}

	for tr.child < len(tr.val.children) {
		if tr.childReader == nil {
			tr.childReader = tr.val.children[tr.child].TokenReader()
		}

		tok, err := tr.childReader.Token()
		if err == io.EOF {
			tr.childReader = nil
			tr.child++
		} else {
			return tok, err
		}
	}

	tr.end = true
	return start.End(), nil
}

var _ xml.TokenReader = (*rawXMLValueReader)(nil)

func valueXMLName(v interface{}) (xml.Name, error) {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return xml.Name{}, fmt.Errorf("webdav: %T is not a struct", v)
	}
	nameField, ok := t.FieldByName("XMLName")
	if !ok {
		return xml.Name{}, fmt.Errorf("webdav: %T is missing an XMLName struct field", v)
	}
	if nameField.Type != reflect.TypeOf(xml.Name{}) {
		return xml.Name{}, fmt.Errorf("webdav: %T.XMLName isn't an xml.Name", v)
	}
	tag := nameField.Tag.Get("xml")
	if tag == "" {
		return xml.Name{}, fmt.Errorf(`webdav: %T.XMLName is missing an "xml" tag`, v)
	}
	name := strings.Split(tag, ",")[0]
	nameParts := strings.Split(name, " ")
	if len(nameParts) != 2 {
		return xml.Name{}, fmt.Errorf("webdav: expected a namespace and local name in %T.XMLName's xml tag", v)
	}
	return xml.Name{nameParts[0], nameParts[1]}, nil
}
